package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

// uidKeys / seqKeys 为常见候选字段名，按优先级顺序探测。
// bot 实测 POST 的 photo dict 字段为 ['imagePath','mysekaiPhotoDecorationId','obtainedAt','seq']
// —— 没有 userId。最可能的是 mysekaiPhotoDecorationId 作为上游 path 第一段。
var (
	uidKeys = []string{
		"mysekaiPhotoDecorationId", "mysekai_photo_decoration_id",
		"userMysekaiPhotoId", "mysekaiPhotoId", "photoId",
		"userId", "user_id", "uid", "id",
	}
	seqKeys = []string{
		"seq", "sequenceId", "sequence_id", "photoIndex",
	}
)

// handlePhoto: GET/POST /api/{region}/mysekai/photo
// body: photo dict (来自 bot mysekai_info.updatedResources.userMysekaiPhotos[i])
// 返回: PNG bytes
func (s *Server) handlePhoto(w http.ResponseWriter, r *http.Request, region string) {
	bodyBytes, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		writeDetailMsg(w, http.StatusBadRequest, "read body failed: "+err.Error())
		return
	}
	var photo map[string]any
	if err := json.Unmarshal(bodyBytes, &photo); err != nil {
		writeDetailMsg(w, http.StatusBadRequest, "invalid photo json: "+err.Error())
		return
	}

	uid, ok := pickStringField(photo, uidKeys)
	if !ok {
		writeDetail(w, http.StatusBadRequest, map[string]any{
			"msg":   "missing uid field in photo body",
			"tried": uidKeys,
			"photo": photo, // 透传整张 dict 便于排查上游真实参数
		})
		return
	}
	seq, ok := pickStringField(photo, seqKeys)
	if !ok {
		writeDetail(w, http.StatusBadRequest, map[string]any{
			"msg":   "missing seq field in photo body",
			"tried": seqKeys,
			"photo": photo,
		})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), time.Duration(s.cfg.HTTP.TimeoutSeconds+2)*time.Second)
	defer cancel()

	res, err := Hedged(ctx, s.upstreams, func(u *Upstream) (string, string, io.Reader, error) {
		base := u.cfg.ImageBase
		if base == "" {
			return "", "", nil, fmt.Errorf("%s missing image_base", u.Name())
		}
		url := fmt.Sprintf("%s/image/%s/mysekai/%s/%s", base, region, uid, seq)
		return http.MethodGet, url, nil, nil
	})
	if err != nil {
		writeDetailMsg(w, http.StatusBadGateway, fmt.Sprintf("upstream failed: %v", err))
		return
	}

	ct := res.ContentType
	if ct == "" {
		ct = "image/png"
	}
	w.Header().Set("Content-Type", ct)
	w.Header().Set("X-Source", res.Source)
	_, _ = w.Write(res.Body)
}

// pickStringField 在 dict 中按顺序找候选字段，返回字符串值。支持 string/number。
func pickStringField(m map[string]any, keys []string) (string, bool) {
	for _, k := range keys {
		v, ok := m[k]
		if !ok {
			continue
		}
		switch x := v.(type) {
		case string:
			if x != "" {
				return x, true
			}
		case float64:
			return strconv.FormatInt(int64(x), 10), true
		case json.Number:
			return x.String(), true
		case int:
			return strconv.Itoa(x), true
		case int64:
			return strconv.FormatInt(x, 10), true
		}
	}
	return "", false
}
