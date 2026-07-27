package http

import "testing"

func TestReportValidation(t *testing.T) {
	if !(reportReq{TargetType: "post", TargetID: "507f1f77bcf86cd799439011", Reason: "Scam link"}).valid() {
		t.Fatal("valid report rejected")
	}
	if (reportReq{TargetType: "profile", TargetID: "bad", Reason: "x"}).valid() {
		t.Fatal("invalid report accepted")
	}
	if !(moderateReq{Status: "resolved"}).valid() {
		t.Fatal("valid moderation status rejected")
	}
}
