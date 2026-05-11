package main

import "net/http"

// 留口的接口，命中后返 501，便于后续逐个补实现。上传 / 数据查询 / 订阅相关由
// sekai-upload 服务承担，不在本网关范畴。
//
// TODO send_boost            : POST /api/{region}/user/{uid}/send_boost  -> {ok_times, failed_reason?}
// TODO create_account        : POST /api/{region}/create_account         -> {inherit_id, inherit_pw}
// TODO ad_result             : GET  /api/{region}/user/{uid}/ad_result   -> {time, results}
// TODO ad_result_update_time : GET  /api/{region}/ad_result/update_time  -> {uid: ts, ...}

func notImplemented(w http.ResponseWriter, name string) {
	writeDetail(w, http.StatusNotImplemented, map[string]string{
		"msg":  "not implemented yet",
		"name": name,
	})
}

func (s *Server) handleStubSendBoost(w http.ResponseWriter, _ *http.Request, _, _ string) {
	notImplemented(w, "send_boost_api_url")
}
func (s *Server) handleStubAdResult(w http.ResponseWriter, _ *http.Request, _, _ string) {
	notImplemented(w, "ad_result_api_url")
}
func (s *Server) handleStubCreateAccount(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "create_account_api_url")
}
func (s *Server) handleStubAdResultUpdateTime(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "ad_result_update_time_api_url")
}
