package main

import "net/http"

// 这些接口与 bot gameapi.py 对应字段一一对应，目前都返回 501，便于后续逐个实现。
//
// TODO suite_api_url           : GET  /api/{region}/user/{uid}/suite
// TODO mysekai_api_url          : GET  /api/{region}/user/{uid}/mysekai
// TODO mysekai_upload_time      : POST /api/{region}/mysekai/upload_time         body=[[uid,mode],...] -> [int,...]
// TODO update_msr_sub           : PUT  /api/{region}/mysekai/subscriptions        body=[[uid,mode],...]
// TODO send_boost               : POST /api/{region}/user/{uid}/send_boost        -> {ok_times, failed_reason?}
// TODO create_account           : POST /api/{region}/create_account               -> {inherit_id, inherit_pw}
// TODO ad_result                : GET  /api/{region}/user/{uid}/ad_result         -> {time, results}
// TODO ad_result_update_time    : GET  /api/{region}/ad_result/update_time        -> {uid: ts, ...}

func notImplemented(w http.ResponseWriter, name string) {
	writeDetail(w, http.StatusNotImplemented, map[string]string{
		"msg":  "not implemented yet",
		"name": name,
	})
}

func (s *Server) handleStubSuite(w http.ResponseWriter, _ *http.Request, _, _ string) {
	notImplemented(w, "suite_api_url")
}
func (s *Server) handleStubMysekai(w http.ResponseWriter, _ *http.Request, _, _ string) {
	notImplemented(w, "mysekai_api_url")
}
func (s *Server) handleStubSendBoost(w http.ResponseWriter, _ *http.Request, _, _ string) {
	notImplemented(w, "send_boost_api_url")
}
func (s *Server) handleStubAdResult(w http.ResponseWriter, _ *http.Request, _, _ string) {
	notImplemented(w, "ad_result_api_url")
}
func (s *Server) handleStubUploadTime(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "mysekai_upload_time_api_url")
}
func (s *Server) handleStubSubscriptions(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "update_msr_sub_api_url")
}
func (s *Server) handleStubCreateAccount(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "create_account_api_url")
}
func (s *Server) handleStubAdResultUpdateTime(w http.ResponseWriter, _ *http.Request, _ string) {
	notImplemented(w, "ad_result_update_time_api_url")
}
