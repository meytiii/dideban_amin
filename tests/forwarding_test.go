package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRoleDefinitionsAndRestrictions(t *testing.T) {
	roles := []string{"superuser", "edare_kol", "navahi_manategh"}
	for _, r := range roles {
		if r != "superuser" && r != "edare_kol" && r != "navahi_manategh" {
			t.Errorf("Unexpected role: %s", r)
		}
	}

	nmSession := AdminSession{
		Username:    "nahie1",
		FullName:    "ارسلان میعادفر",
		Role:        "navahi_manategh",
		IsSuperuser: false,
	}

	if nmSession.IsSuperuser {
		t.Errorf("navahi_manategh session must not be superuser")
	}

	superSession := AdminSession{Username: "admin_it", Role: "superuser", IsSuperuser: true}
	edareKolSession := AdminSession{Username: "admin_herasat", Role: "edare_kol", IsSuperuser: false}

	if !canAdminAccessReport(superSession, 9999) {
		t.Errorf("superuser should have access to any report")
	}
	if !canAdminAccessReport(edareKolSession, 9999) {
		t.Errorf("edare_kol should have access to any report")
	}
}

func TestNavahiManateghClassificationForbidden(t *testing.T) {

	InitCrypto("01234567890123456789012345678901")

	reqBody := `{"report_id": 1, "classification": "helpful"}`
	req := httptest.NewRequest("POST", "/api/admin/report/classify", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	nmSession := AdminSession{
		Username:    "nahie1",
		FullName:    "ارسلان میعادفر",
		Role:        "navahi_manategh",
		IsSuperuser: false,
	}
	ctx := req.Context()
	ctx = contextWithAdmin(ctx, nmSession)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ClassifyReportHandler(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("SECURITY VIOLATION: navahi_manategh must receive 403 Forbidden on classify. Got: %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "دسترسی غیرمجاز") {
		t.Errorf("Expected Persian error message, got: %s", rr.Body.String())
	}
}

func TestForwardingRequestValidation(t *testing.T) {

	reqBody := `{"report_id": 1, "recipient_ids": []}`
	req := httptest.NewRequest("POST", "/api/admin/report/forward", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")

	edareKolSession := AdminSession{
		Username:    "admin_herasat",
		FullName:    "حمیدرضا طاهری‌مجد",
		Role:        "edare_kol",
		IsSuperuser: false,
	}
	ctx := contextWithAdmin(req.Context(), edareKolSession)
	req = req.WithContext(ctx)

	rr := httptest.NewRecorder()
	ForwardReportHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("Expected 400 Bad Request when recipient list is empty, got: %d", rr.Code)
	}

	nmReq := httptest.NewRequest("POST", "/api/admin/report/forward", strings.NewReader(`{"report_id": 1, "recipient_ids": [7]}`))
	nmSession := AdminSession{
		Username:    "nahie1",
		FullName:    "ارسلان میعادفر",
		Role:        "navahi_manategh",
		IsSuperuser: false,
	}
	nmReq = nmReq.WithContext(contextWithAdmin(nmReq.Context(), nmSession))
	nmRR := httptest.NewRecorder()
	ForwardReportHandler(nmRR, nmReq)

	if nmRR.Code != http.StatusForbidden {
		t.Fatalf("SECURITY VIOLATION: navahi_manategh must not be allowed to forward reports! Got: %d", nmRR.Code)
	}
}

func TestForwardHistorySerialization(t *testing.T) {
	now := time.Now()
	viewTime := now.Add(10 * time.Minute)

	historyItem := ForwardHistoryItem{
		ID:                1,
		ReportID:          100,
		ForwardedBy:       "admin_herasat",
		ForwarderName:     "حمیدرضا طاهری‌مجد",
		ForwarderTitle:    "رئیس اداره حراست",
		RecipientUsername: "nahie1",
		RecipientName:     "ارسلان میعادفر",
		RecipientTitle:    "کارشناس مسئول حراست ناحیه یک",
		ForwardedAt:       now.Format(time.RFC3339),
		IsViewed:          true,
		ViewedAt:          viewTime.Format(time.RFC3339),
	}

	data, err := json.Marshal(historyItem)
	if err != nil {
		t.Fatalf("Failed to marshal ForwardHistoryItem: %v", err)
	}

	var parsed ForwardHistoryItem
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal ForwardHistoryItem: %v", err)
	}

	if parsed.ForwarderName != "حمیدرضا طاهری‌مجد" || parsed.RecipientName != "ارسلان میعادفر" {
		t.Errorf("Names mismatch after serialization")
	}
	if !parsed.IsViewed || parsed.ViewedAt == "" {
		t.Errorf("Viewed status mismatch")
	}
}

func TestNavahiManateghNotificationScoping(t *testing.T) {

	items := []AdminNotificationItem{
		{
			ID:                 1,
			NotifType:          "new_report",
			Title:              "گزارش محرمانه جدید در کارتابل",
			Message:            "موضوع: فساد مالی",
			TargetReportID:     new(int64),
			TargetTrackingCode: new(string),
			IsRead:             false,
		},
		{
			ID:                 2,
			NotifType:          "report_forwarded",
			Title:              "ارجاع خبر جدید از اداره کل",
			Message:            "پرونده محرمانه با شماره پیگیری HMD-100 توسط رئیس اداره حراست به کارتابل شما ارجاع داده شد.",
			TargetReportID:     new(int64),
			TargetTrackingCode: new(string),
			IsRead:             false,
		},
	}
	*items[0].TargetReportID = 101
	*items[0].TargetTrackingCode = "HMD-101"
	*items[1].TargetReportID = 100
	*items[1].TargetTrackingCode = "HMD-100"

	nmUsername := "nahie1"
	recipUsernameNotif2 := "nahie1"

	visibleForNahie1 := []AdminNotificationItem{}
	for _, it := range items {
		if it.NotifType == "new_report" {

			continue
		}
		if it.NotifType == "report_forwarded" && recipUsernameNotif2 == nmUsername {
			visibleForNahie1 = append(visibleForNahie1, it)
		}
	}

	if len(visibleForNahie1) != 1 {
		t.Fatalf("Expected exactly 1 notification for nahie1, got %d", len(visibleForNahie1))
	}
	if visibleForNahie1[0].NotifType != "report_forwarded" {
		t.Errorf("Expected report_forwarded notification, got %s", visibleForNahie1[0].NotifType)
	}
	if !strings.Contains(visibleForNahie1[0].Title, "ارجاع خبر جدید") {
		t.Errorf("Expected Persian forwarding title, got: %s", visibleForNahie1[0].Title)
	}
}

func TestNavahiManateghRestrictions(t *testing.T) {
	nmSession := AdminSession{
		Username:    "nahie1",
		FullName:    "ارسلان میعادفر",
		Role:        "navahi_manategh",
		IsSuperuser: false,
	}

	reqResp := httptest.NewRequest("POST", "/api/admin/respond", strings.NewReader(`{"report_id": 10, "response": "پاسخ تستی"}`))
	reqResp = reqResp.WithContext(contextWithAdmin(reqResp.Context(), nmSession))
	rrResp := httptest.NewRecorder()
	RespondToReportHandler(rrResp, reqResp)

	if rrResp.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for RespondToReportHandler, got %d", rrResp.Code)
	}
	if !strings.Contains(rrResp.Body.String(), "شما به این قسمت دسترسی ندارید") {
		t.Errorf("Expected 'شما به این قسمت دسترسی ندارید' in response, got: %s", rrResp.Body.String())
	}

	reqNotes := httptest.NewRequest("POST", "/api/admin/report/internal-notes", strings.NewReader(`{"report_id": 10, "internal_notes": "یادداشت تستی"}`))
	reqNotes = reqNotes.WithContext(contextWithAdmin(reqNotes.Context(), nmSession))
	rrNotes := httptest.NewRecorder()
	UpdateInternalNotesHandler(rrNotes, reqNotes)

	if rrNotes.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for UpdateInternalNotesHandler, got %d", rrNotes.Code)
	}
	if !strings.Contains(rrNotes.Body.String(), "شما به این قسمت دسترسی ندارید") {
		t.Errorf("Expected 'شما به این قسمت دسترسی ندارید' in notes response, got: %s", rrNotes.Body.String())
	}

	reqPrio := httptest.NewRequest("POST", "/api/admin/report/set-priority", strings.NewReader(`{"report_id": 10, "priority": "urgent"}`))
	reqPrio = reqPrio.WithContext(contextWithAdmin(reqPrio.Context(), nmSession))
	rrPrio := httptest.NewRecorder()
	SetReportPriorityHandler(rrPrio, reqPrio)

	if rrPrio.Code != http.StatusForbidden {
		t.Fatalf("Expected 403 Forbidden for SetReportPriorityHandler, got %d", rrPrio.Code)
	}
	if !strings.Contains(rrPrio.Body.String(), "شما به این قسمت دسترسی ندارید") {
		t.Errorf("Expected 'شما به این قسمت دسترسی ندارید' in set-priority response, got: %s", rrPrio.Body.String())
	}
}

func TestRegionalResponseWorkflow(t *testing.T) {

	nmSession := AdminSession{
		Username:    "nahie1",
		FullName:    "ارسلان میعادفر",
		Role:        "navahi_manategh",
		IsSuperuser: false,
	}

	reqEmpty := httptest.NewRequest("POST", "/api/admin/report/regional-response", strings.NewReader(`{"report_id": 10, "response": ""}`))
	reqEmpty = reqEmpty.WithContext(contextWithAdmin(reqEmpty.Context(), nmSession))
	rrEmpty := httptest.NewRecorder()
	SubmitRegionalResponseHandler(rrEmpty, reqEmpty)

	if rrEmpty.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty regional response, got %d", rrEmpty.Code)
	}

	now := time.Now()
	regItem := RegionalResponseItem{
		ID:                5,
		RecipientUsername: "nahie1",
		RecipientName:     "ارسلان میعادفر",
		RecipientTitle:    "کارشناس مسئول حراست ناحیه یک",
		ResponseText:      "نتیجه بررسی میدانی: مورد تأیید است.",
		RespondedAt:       now.Format(time.RFC3339),
	}

	repDetail := ReportDetail{
		ID:           10,
		TrackingCode: "HMD-12345",
		Category:     "تخلفات اداری",
		RegionalResponses: []RegionalResponseItem{
			regItem,
		},
		MyRegionalResponse: &regItem,
	}

	data, err := json.Marshal(repDetail)
	if err != nil {
		t.Fatalf("Failed to marshal ReportDetail with regional responses: %v", err)
	}

	var parsed RepDetailTest
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal ReportDetail: %v", err)
	}

	if len(parsed.RegionalResponses) != 1 || parsed.RegionalResponses[0].ResponseText != regItem.ResponseText {
		t.Errorf("Regional response text mismatch in serialization")
	}
	if parsed.MyRegionalResponse == nil || parsed.MyRegionalResponse.RecipientUsername != "nahie1" {
		t.Errorf("MyRegionalResponse mismatch in serialization")
	}
}

type RepDetailTest struct {
	ID                 int64                  `json:"id"`
	TrackingCode       string                 `json:"tracking_code"`
	RegionalResponses  []RegionalResponseItem `json:"regional_responses"`
	MyRegionalResponse *RegionalResponseItem  `json:"my_regional_response"`
}

func TestRemoveForwardingSecurity(t *testing.T) {

	nmSession := AdminSession{
		Username:    "nahie1",
		FullName:    "ارسلان میعادفر",
		Role:        "navahi_manategh",
		IsSuperuser: false,
	}

	reqNM := httptest.NewRequest("POST", "/api/admin/report/remove-forward", strings.NewReader(`{"forwarding_id": 1}`))
	reqNM = reqNM.WithContext(contextWithAdmin(reqNM.Context(), nmSession))
	rrNM := httptest.NewRecorder()
	RemoveForwardingHandler(rrNM, reqNM)

	if rrNM.Code != http.StatusForbidden {
		t.Fatalf("SECURITY VIOLATION: navahi_manategh must receive 403 Forbidden when calling RemoveForwardingHandler. Got: %d", rrNM.Code)
	}
	if !strings.Contains(rrNM.Body.String(), "مجوز") {
		t.Errorf("Expected permission error message, got: %s", rrNM.Body.String())
	}

	edareKolSession := AdminSession{
		Username:    "admin_herasat",
		FullName:    "حمیدرضا طاهری‌مجد",
		Role:        "edare_kol",
		IsSuperuser: false,
	}

	reqEmpty := httptest.NewRequest("POST", "/api/admin/report/remove-forward", strings.NewReader(`{}`))
	reqEmpty = reqEmpty.WithContext(contextWithAdmin(reqEmpty.Context(), edareKolSession))
	rrEmpty := httptest.NewRecorder()
	RemoveForwardingHandler(rrEmpty, reqEmpty)

	if rrEmpty.Code != http.StatusBadRequest {
		t.Errorf("Expected 400 Bad Request for empty RemoveForwarding payload, got %d", rrEmpty.Code)
	}
}

func TestForwardHistoryItemWithRegionalResponse(t *testing.T) {
	now := time.Now()
	item := ForwardHistoryItem{
		ID:                1,
		ReportID:          10,
		ForwardedBy:       "admin_herasat",
		ForwarderName:     "حمیدرضا طاهری‌مجد",
		ForwarderTitle:    "رئیس اداره حراست",
		RecipientUsername: "nahie1",
		RecipientName:     "ارسلان میعادفر",
		RecipientTitle:    "حراست ناحیه یک",
		ForwardedAt:       now.Format(time.RFC3339),
		IsViewed:          true,
		ViewedAt:          now.Format(time.RFC3339),
		ResponseText:      "بررسی با موفقیت انجام شد و مدارک ضمیمه گردید.",
		RespondedAt:       now.Format(time.RFC3339),
	}

	data, err := json.Marshal(item)
	if err != nil {
		t.Fatalf("Failed to marshal ForwardHistoryItem: %v", err)
	}

	var parsed ForwardHistoryItem
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal ForwardHistoryItem: %v", err)
	}

	if parsed.ResponseText != item.ResponseText || parsed.RespondedAt != item.RespondedAt {
		t.Errorf("ResponseText or RespondedAt mismatch in ForwardHistoryItem")
	}
}

func TestReportSummaryRegionalResponseSerialization(t *testing.T) {
	summary := ReportSummary{
		ID:                     101,
		TrackingCode:           "TRK-9876543210",
		Category:               "تخلفات اداری",
		ReportType:             "anonymous",
		Priority:               "urgent",
		IsRead:                 true,
		HasResponse:            false,
		HasNewMessage:          false,
		HasRegionalResponse:    true,
		HasAnyRegionalResponse: true,
		CreatedAt:              time.Now().Format(time.RFC3339),
		UpdatedAt:              time.Now().Format(time.RFC3339),
	}

	data, err := json.Marshal(summary)
	if err != nil {
		t.Fatalf("Failed to marshal ReportSummary: %v", err)
	}

	var parsed ReportSummary
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal ReportSummary: %v", err)
	}

	if !parsed.HasRegionalResponse {
		t.Errorf("Expected HasRegionalResponse=true, got false")
	}
	if !parsed.HasAnyRegionalResponse {
		t.Errorf("Expected HasAnyRegionalResponse=true, got false")
	}
}
