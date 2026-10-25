package main

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync/atomic"
	"time"
	"unicode/utf8"

	"github.com/lib/pq"
)

func AdminAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AdminIPWhitelistMiddleware(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
		w.Header().Set("Pragma", "no-cache")
		w.Header().Set("X-Server-Time", time.Now().Format(time.RFC3339Nano))

		cookie, err := r.Cookie("herasat_admin_session")
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		session, exists := Sessions.Get(cookie.Value)
		if !exists {
			if reason, wasKicked := Sessions.GetKickedReason(cookie.Value); wasKicked {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{
					"error":   "concurrent_login",
					"reason":  reason,
					"message": "یک ورود جدید با حساب کاربری شما در دستگاه دیگری انجام شد و شما از این سامانه خارج شدید.",
				})
				return
			}
			http.Error(w, "Session expired or invalid", http.StatusUnauthorized)
			return
		}

		Sessions.UpdateActivity(cookie.Value)
		session, _ = Sessions.Get(cookie.Value)

		http.SetCookie(w, &http.Cookie{
			Name:     "herasat_admin_session",
			Value:    cookie.Value,
			Path:     "/",
			Expires:  session.Expiry,
			HttpOnly: true,
			Secure:   isCookieSecure(),
			SameSite: http.SameSiteLaxMode,
		})

		var currentCSRF string
		if csrfCookie, err := r.Cookie("herasat_csrf"); err == nil && csrfCookie.Value != "" {
			currentCSRF = csrfCookie.Value
		} else {
			csrfBytes := make([]byte, 32)
			rand.Read(csrfBytes)
			currentCSRF = base64.RawURLEncoding.EncodeToString(csrfBytes)
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "herasat_csrf",
			Value:    currentCSRF,
			Path:     "/",
			Expires:  session.Expiry,
			HttpOnly: false,
			Secure:   isCookieSecure(),
			SameSite: http.SameSiteStrictMode,
		})
		w.Header().Set("X-CSRF-Token", currentCSRF)

		ctx := context.WithValue(r.Context(), "admin", session)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func SuperuserAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return AdminAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
		admin := r.Context().Value("admin").(AdminSession)
		if !admin.IsSuperuser && admin.Role != "superuser" {
			http.Error(w, "دسترسی غیرمجاز: این بخش منحصراً متعلق به پشتیبانی سامانه است.", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func canAdminAccessReport(admin AdminSession, reportID int64) bool {
	if admin.IsSuperuser || admin.Role == "superuser" || admin.Role == "edare_kol" {
		return true
	}
	if admin.Role == "navahi_manategh" {
		var exists bool
		err := DB.QueryRow(`
			SELECT EXISTS(
				SELECT 1 FROM report_forwardings
				WHERE report_id = $1 AND recipient_username = $2
			)
		`, reportID, admin.Username).Scan(&exists)
		return err == nil && exists
	}
	return false
}

func ToggleReadStatusHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		ReportID int64 `json:"report_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, req.ReportID) {
		http.Error(w, "دسترسی غیرمجاز: شما اجازه بررسی این پرونده را ندارید.", http.StatusForbidden)
		return
	}

	var isViewed bool
	err := DB.QueryRow(`
		SELECT EXISTS(
			SELECT 1 FROM admin_report_views
			WHERE admin_username = $1 AND report_id = $2
		)
	`, admin.Username, req.ReportID).Scan(&isViewed)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if isViewed {

		_, err = DB.Exec(`
			DELETE FROM admin_report_views
			WHERE admin_username = $1 AND report_id = $2
		`, admin.Username, req.ReportID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		logAudit(admin.Username, admin.Title, "MARK_UNREAD", &req.ReportID, getIP(r))
	} else {

		_, err = DB.Exec(`
			INSERT INTO admin_report_views (admin_username, report_id, viewed_at, last_viewed_message_id)
			VALUES ($1, $2, NOW(), COALESCE((SELECT MAX(id) FROM report_messages WHERE report_id = $2), 0))
			ON CONFLICT (admin_username, report_id)
			DO UPDATE SET viewed_at = NOW(),
			              last_viewed_message_id = COALESCE((SELECT MAX(id) FROM report_messages WHERE report_id = $2), 0)
		`, admin.Username, req.ReportID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		if admin.Role == "navahi_manategh" {
			_, _ = DB.Exec(`
				UPDATE admin_notifications
				SET is_read = true
				WHERE target_report_id = $1 AND is_read = false AND recipient_username = $2
			`, req.ReportID, admin.Username)
		} else {
			_, _ = DB.Exec(`
				UPDATE admin_notifications
				SET is_read = true
				WHERE target_report_id = $1 AND is_read = false AND (recipient_username IS NULL OR recipient_username = $2)
			`, req.ReportID, admin.Username)
		}

		logAudit(admin.Username, admin.Title, "MARK_READ", &req.ReportID, getIP(r))
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success"}`))
}

type ReportSummary struct {
	ID                     int64  `json:"id"`
	TrackingCode           string `json:"tracking_code"`
	Category               string `json:"category"`
	ReportType             string `json:"report_type"`
	Priority               string `json:"priority"`
	IsRead                 bool   `json:"is_read"`
	HasResponse            bool   `json:"has_response"`
	HasNewMessage          bool   `json:"has_new_message"`
	HasRegionalResponse    bool   `json:"has_regional_response"`
	HasAnyRegionalResponse bool   `json:"has_any_regional_response,omitempty"`
	CreatedAt              string `json:"created_at"`
	UpdatedAt              string `json:"updated_at"`
	ReadByName             string `json:"read_by_name"`
	Classification         string `json:"classification,omitempty"`
	Description            string `json:"description,omitempty"`
}

type PaginatedReportsResponse struct {
	Reports  []ReportSummary `json:"reports"`
	Total    int64           `json:"total"`
	Page     int             `json:"page"`
	PageSize int             `json:"page_size"`
}

func normalizeForSearch(s string) string {
	lowered := strings.ToLower(s)
	var b strings.Builder
	for _, r := range lowered {
		if isDiacriticOrInvisible(r) {
			continue
		}
		r = cleanRune(r)
		if r == 'آ' || r == 'أ' || r == 'إ' {
			r = 'ا'
		}
		b.WriteRune(r)
	}
	return b.String()
}

func GetReportsHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	page := 1
	pageSize := 25
	fetchAll := r.URL.Query().Get("all") == "true" || r.URL.Query().Get("page_size") == "all"
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" && ps != "all" {
		if parsed, err := strconv.Atoi(ps); err == nil && parsed > 0 && parsed <= 10000 {
			pageSize = parsed
		}
	}
	if fetchAll {
		pageSize = 100000
		page = 1
	}

	offset := (page - 1) * pageSize

	var whereClauses []string
	var args []interface{}

	args = append(args, admin.Username)

	if admin.Role == "navahi_manategh" {
		whereClauses = append(whereClauses, "rf.recipient_username = $1")
	} else {
		whereClauses = append(whereClauses, "1=1")
	}

	searchQuery := strings.TrimSpace(r.URL.Query().Get("q"))
	if searchQuery == "" {
		searchQuery = strings.TrimSpace(r.URL.Query().Get("search"))
	}
	if searchQuery != "" {
		args = append(args, "%"+searchQuery+"%")
		paramIdx := len(args)
		whereClauses = append(whereClauses, fmt.Sprintf("(r.tracking_code ILIKE $%d OR r.category ILIKE $%d)", paramIdx, paramIdx))
	}

	category := strings.TrimSpace(r.URL.Query().Get("category"))
	if category != "" && category != "all" {
		args = append(args, category)
		whereClauses = append(whereClauses, fmt.Sprintf("r.category = $%d", len(args)))
	}

	priority := strings.TrimSpace(r.URL.Query().Get("priority"))
	if priority != "" && priority != "all" {
		args = append(args, priority)
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(r.priority, 'normal') = $%d", len(args)))
	}

	classification := strings.TrimSpace(r.URL.Query().Get("classification"))
	if classification != "" && classification != "all" {
		args = append(args, classification)
		whereClauses = append(whereClauses, fmt.Sprintf("COALESCE(r.classification, '') = $%d", len(args)))
	}

	readStatus := strings.TrimSpace(r.URL.Query().Get("read_status"))
	if readStatus != "" && readStatus != "all" {
		if admin.Role == "navahi_manategh" {
			switch readStatus {
			case "unread":
				whereClauses = append(whereClauses, "arv.id IS NULL")
			case "read":
				whereClauses = append(whereClauses, "arv.id IS NOT NULL AND (rf.response_text IS NULL OR rf.response_text = '') AND NOT EXISTS (SELECT 1 FROM report_messages rm WHERE rm.report_id = r.id AND rm.sender_type = 'reporter' AND rm.id > COALESCE(arv.last_viewed_message_id, 0))")
			case "responded":
				whereClauses = append(whereClauses, "rf.response_text IS NOT NULL AND rf.response_text != ''")
			case "new_message":
				whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM report_messages rm WHERE rm.report_id = r.id AND rm.sender_type = 'reporter' AND rm.id > COALESCE(arv.last_viewed_message_id, 0))")
			}
		} else {
			switch readStatus {
			case "unread":
				whereClauses = append(whereClauses, "arv.id IS NULL")
			case "read":
				whereClauses = append(whereClauses, "arv.id IS NOT NULL AND (r.admin_response IS NULL OR r.admin_response = '') AND NOT EXISTS (SELECT 1 FROM report_messages rm WHERE rm.report_id = r.id AND rm.sender_type = 'reporter' AND rm.id > COALESCE(arv.last_viewed_message_id, 0)) AND NOT EXISTS (SELECT 1 FROM report_forwardings rf2 WHERE rf2.report_id = r.id AND rf2.response_text IS NOT NULL AND rf2.response_text != '')")
			case "responded":
				whereClauses = append(whereClauses, "r.admin_response IS NOT NULL AND r.admin_response != ''")
			case "new_message":
				whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM report_messages rm WHERE rm.report_id = r.id AND rm.sender_type = 'reporter' AND rm.id > COALESCE(arv.last_viewed_message_id, 0))")
			case "regional_response":
				whereClauses = append(whereClauses, "EXISTS (SELECT 1 FROM report_forwardings rf2 WHERE rf2.report_id = r.id AND rf2.response_text IS NOT NULL AND rf2.response_text != '')")
			}
		}
	}

	whereSql := strings.Join(whereClauses, " AND ")

	countArgs := make([]interface{}, len(args))
	copy(countArgs, args)

	var total int64
	var countQuery string
	if admin.Role == "navahi_manategh" {
		countQuery = `
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			LEFT JOIN admin_report_views arv ON r.id = arv.report_id AND arv.admin_username = $1
			WHERE ` + whereSql
	} else {
		countQuery = `
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			LEFT JOIN admin_report_views arv ON r.id = arv.report_id AND arv.admin_username = $1
			WHERE ` + whereSql
	}

	if err := DB.QueryRow(countQuery, countArgs...).Scan(&total); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	args = append(args, pageSize, offset)
	limitIdx := len(args) - 1
	offsetIdx := len(args)

	var selectQuery string
	if admin.Role == "navahi_manategh" {
		selectQuery = `
			SELECT r.id, r.tracking_code, r.category, r.report_type, COALESCE(r.priority, 'normal'),
			       (arv.id IS NOT NULL) AS is_read,
			       (rf.response_text IS NOT NULL AND rf.response_text != '') AS has_response, r.created_at,
			       EXISTS(
			           SELECT 1 FROM report_messages rm
			           WHERE rm.report_id = r.id AND rm.sender_type = 'reporter'
			             AND rm.id > COALESCE(arv.last_viewed_message_id, 0)
			       ) AS has_new_message,
			       FALSE AS has_regional_response,
			       FALSE AS has_any_regional_response,
			       r.updated_at,
			       CASE WHEN arv.id IS NOT NULL THEN 'شما' ELSE '' END,
			       COALESCE(r.classification, ''),
			       r.description
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			LEFT JOIN admin_report_views arv ON r.id = arv.report_id AND arv.admin_username = $1
			WHERE ` + whereSql + `
			ORDER BY r.created_at DESC, r.id DESC
			LIMIT $` + strconv.Itoa(limitIdx) + ` OFFSET $` + strconv.Itoa(offsetIdx)
	} else {
		selectQuery = `
			SELECT r.id, r.tracking_code, r.category, r.report_type, COALESCE(r.priority, 'normal'),
			       (arv.id IS NOT NULL) AS is_read,
			       (r.admin_response IS NOT NULL AND r.admin_response != '') AS has_response, r.created_at,
			       EXISTS(
			           SELECT 1 FROM report_messages rm
			           WHERE rm.report_id = r.id AND rm.sender_type = 'reporter'
			             AND rm.id > COALESCE(arv.last_viewed_message_id, 0)
			       ) AS has_new_message,
			       EXISTS(
			           SELECT 1 FROM report_forwardings rf
			           WHERE rf.report_id = r.id
			             AND rf.response_text IS NOT NULL
			             AND rf.response_text != ''
			             AND (arv.viewed_at IS NULL OR COALESCE(rf.responded_at, rf.created_at) > arv.viewed_at)
			       ) AS has_regional_response,
			       EXISTS(
			           SELECT 1 FROM report_forwardings rf
			           WHERE rf.report_id = r.id
			             AND rf.response_text IS NOT NULL
			             AND rf.response_text != ''
			       ) AS has_any_regional_response,
			       r.updated_at,
			       CASE WHEN arv.id IS NOT NULL THEN 'شما' ELSE '' END,
			       COALESCE(r.classification, ''),
			       r.description
			FROM reports r
			LEFT JOIN admin_report_views arv ON r.id = arv.report_id AND arv.admin_username = $1
			WHERE ` + whereSql + `
			ORDER BY r.created_at DESC, r.id DESC
			LIMIT $` + strconv.Itoa(limitIdx) + ` OFFSET $` + strconv.Itoa(offsetIdx)
	}

	rows, err := DB.Query(selectQuery, args...)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var reports []ReportSummary
	for rows.Next() {
		var rep ReportSummary
		var createdAt, updatedAt time.Time
		var readByName, classification, encDesc sql.NullString
		if err := rows.Scan(&rep.ID, &rep.TrackingCode, &rep.Category, &rep.ReportType, &rep.Priority, &rep.IsRead, &rep.HasResponse, &createdAt, &rep.HasNewMessage, &rep.HasRegionalResponse, &rep.HasAnyRegionalResponse, &updatedAt, &readByName, &classification, &encDesc); err != nil {
			continue
		}
		rep.CreatedAt = createdAt.Format(time.RFC3339)
		rep.UpdatedAt = updatedAt.Format(time.RFC3339)
		if readByName.Valid {
			rep.ReadByName = readByName.String
		}
		if classification.Valid {
			rep.Classification = classification.String
		}
		if encDesc.Valid && encDesc.String != "" {
			if decDesc, err := DecryptAES(encDesc.String); err == nil {
				rep.Description = decDesc
			} else {
				rep.Description = "[خطای رمزگشایی - اصالت محتوا تأیید نشد]"
			}
		}
		reports = append(reports, rep)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if reports == nil {
		reports = []ReportSummary{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedReportsResponse{
		Reports:  reports,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

type RegionalResponseItem struct {
	ID                int64  `json:"id"`
	RecipientUsername string `json:"recipient_username"`
	RecipientName     string `json:"recipient_name"`
	RecipientTitle    string `json:"recipient_title"`
	ResponseText      string `json:"response_text"`
	RespondedAt       string `json:"responded_at"`
}

type ReportDetail struct {
	ID                 int64                  `json:"id"`
	TrackingCode       string                 `json:"tracking_code"`
	Category           string                 `json:"category"`
	ReportType         string                 `json:"report_type"`
	Priority           string                 `json:"priority"`
	ReporterName       string                 `json:"reporter_name"`
	ReporterPhone      string                 `json:"reporter_phone"`
	Description        string                 `json:"description"`
	InternalNotes      string                 `json:"internal_notes"`
	IsRead             bool                   `json:"is_read"`
	HasNewMessage      bool                   `json:"has_new_message"`
	CreatedAt          string                 `json:"created_at"`
	ReadAt             string                 `json:"read_at"`
	ReadBy             string                 `json:"read_by"`
	ReadByName         string                 `json:"read_by_name"`
	Classification     string                 `json:"classification"`
	ClassifiedBy       string                 `json:"classified_by"`
	ClassifiedByName   string                 `json:"classified_by_name"`
	ClassifiedAt       string                 `json:"classified_at"`
	HasAttachment      bool                   `json:"has_attachment"`
	Attachments        []ReportAttachmentItem `json:"attachments"`
	AdminResponse      string                 `json:"admin_response"`
	RespondedAt        string                 `json:"responded_at"`
	RespondedBy        string                 `json:"responded_by"`
	Messages           []ReportMessageItem    `json:"messages"`
	RegionalResponses  []RegionalResponseItem `json:"regional_responses,omitempty"`
	MyRegionalResponse *RegionalResponseItem  `json:"my_regional_response,omitempty"`
	HasResponse        bool                   `json:"has_response"`
}

type ReportAttachmentItem struct {
	ID       int64  `json:"id"`
	FileName string `json:"file_name"`
	FileSize int64  `json:"file_size"`
	MimeType string `json:"mime_type"`
}

func GetReportDetailsHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	reportIDStr := r.URL.Query().Get("id")
	reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, reportID) {
		http.Error(w, "دسترسی غیرمجاز: شما مجوز دسترسی به این پرونده را ندارید.", http.StatusForbidden)
		return
	}

	if admin.Role == "navahi_manategh" {
		_, _ = DB.Exec(`
			UPDATE report_forwardings
			SET is_viewed = true,
			    viewed_at = COALESCE(viewed_at, NOW())
			WHERE report_id = $1 AND recipient_username = $2 AND is_viewed = false
		`, reportID, admin.Username)
	}

	logAudit(admin.Username, admin.Title, "VIEW_REPORT_DETAILS", &reportID, getIP(r))

	_, _ = DB.Exec(`
		INSERT INTO admin_report_views (admin_username, report_id, viewed_at, last_viewed_message_id)
		VALUES ($1, $2, NOW(), COALESCE((SELECT MAX(id) FROM report_messages WHERE report_id = $2), 0))
		ON CONFLICT (admin_username, report_id)
		DO UPDATE SET viewed_at = NOW(),
		              last_viewed_message_id = COALESCE((SELECT MAX(id) FROM report_messages WHERE report_id = $2), 0)
	`, admin.Username, reportID)

	_, _ = DB.Exec(`
		UPDATE reports
		SET is_read = true,
		    read_at = COALESCE(read_at, NOW()),
		    read_by = COALESCE(read_by, $1),
		    has_new_message = false
		WHERE id = $2
	`, admin.Username, reportID)

	if admin.Role == "navahi_manategh" {
		_, _ = DB.Exec(`
			UPDATE admin_notifications
			SET is_read = true
			WHERE target_report_id = $1 AND is_read = false AND recipient_username = $2
		`, reportID, admin.Username)
	} else {
		_, _ = DB.Exec(`
			UPDATE admin_notifications
			SET is_read = true
			WHERE target_report_id = $1 AND is_read = false AND (recipient_username IS NULL OR recipient_username = $2)
		`, reportID, admin.Username)
	}

	var rep ReportDetail
	var encName, encPhone, encDesc, encResponse, encNotes, respBy, priority sql.NullString
	var readBy, readByName, classification, classifiedBy, classifiedByName sql.NullString
	var createdAt time.Time
	var readAt, respAt, classifiedAt sql.NullTime
	var evidenceID sql.NullInt64
	var hasNewMsg bool

	err = DB.QueryRow(`
		SELECT r.id, r.tracking_code, r.category, r.report_type, COALESCE(r.priority, 'normal'),
		       r.reporter_name, r.reporter_phone, r.description, r.internal_notes,
		       r.is_read, r.created_at, r.read_at, r.admin_response, r.responded_at,
		       COALESCE(NULLIF(u.full_name, ''), r.responded_by, ''), e.id,
		       COALESCE(r.read_by, ''),
		       COALESCE(NULLIF(u_read.full_name, ''), r.read_by, ''),
		       COALESCE(r.has_new_message, false),
		       COALESCE(r.classification, ''),
		       COALESCE(r.classified_by, ''),
		       COALESCE(NULLIF(u_class.full_name, ''), r.classified_by, ''),
		       r.classified_at
		FROM reports r
		LEFT JOIN admin_users u ON r.responded_by = u.username
		LEFT JOIN admin_users u_read ON r.read_by = u_read.username
		LEFT JOIN admin_users u_class ON r.classified_by = u_class.username
		LEFT JOIN evidence_files e ON r.id = e.report_id
		WHERE r.id = $1`, reportID).
		Scan(&rep.ID, &rep.TrackingCode, &rep.Category, &rep.ReportType, &priority,
			&encName, &encPhone, &encDesc, &encNotes,
			&rep.IsRead, &createdAt, &readAt, &encResponse, &respAt, &respBy, &evidenceID,
			&readBy, &readByName, &hasNewMsg,
			&classification, &classifiedBy, &classifiedByName, &classifiedAt)

	if err == sql.ErrNoRows {
		http.Error(w, "Report not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if classification.Valid {
		rep.Classification = classification.String
	}
	if classifiedBy.Valid {
		rep.ClassifiedBy = classifiedBy.String
	}
	if classifiedByName.Valid {
		rep.ClassifiedByName = classifiedByName.String
	}
	if classifiedAt.Valid {
		rep.ClassifiedAt = classifiedAt.Time.Format(time.RFC3339)
	}

	rep.IsRead = true
	rep.HasNewMessage = false
	rep.ReadBy = admin.Username
	rep.ReadByName = "شما"
	rep.ReadAt = time.Now().Format(time.RFC3339)

	if priority.Valid && priority.String != "" {
		rep.Priority = priority.String
	} else {
		rep.Priority = "normal"
	}

	if encName.Valid && encName.String != "" {
		if dec, err := DecryptAES(encName.String); err == nil {
			rep.ReporterName = dec
		} else {
			rep.ReporterName = "[خطای رمزگشایی]"
		}
	} else {
		rep.ReporterName = "ندارد"
	}

	if encPhone.Valid && encPhone.String != "" {
		if dec, err := DecryptAES(encPhone.String); err == nil {
			if strings.TrimSpace(dec) != "" {
				rep.ReporterPhone = dec
			} else {
				rep.ReporterPhone = "ندارد"
			}
		} else {
			rep.ReporterPhone = "[خطای رمزگشایی]"
		}
	} else {
		rep.ReporterPhone = "ندارد"
	}

	if encDesc.Valid && encDesc.String != "" {
		if dec, err := DecryptAES(encDesc.String); err == nil {
			rep.Description = dec
		} else {
			rep.Description = "[خطای رمزگشایی - اصالت محتوا تأیید نشد]"
		}
	} else {
		rep.Description = "ندارد"
	}

	if encNotes.Valid && encNotes.String != "" {
		decNotes, err := DecryptAES(encNotes.String)
		if err == nil {
			rep.InternalNotes = decNotes
		} else {
			rep.InternalNotes = ""
		}
	} else {
		rep.InternalNotes = ""
	}

	if encResponse.Valid && encResponse.String != "" {
		if dec, err := DecryptAES(encResponse.String); err == nil {
			rep.AdminResponse = dec
		} else {
			rep.AdminResponse = "[خطای رمزگشایی]"
		}
	} else {
		rep.AdminResponse = ""
	}

	if respBy.Valid {
		rep.RespondedBy = respBy.String
	}

	rep.CreatedAt = createdAt.Format(time.RFC3339)
	if readAt.Valid {
		rep.ReadAt = readAt.Time.Format(time.RFC3339)
	} else {
		rep.ReadAt = ""
	}

	if respAt.Valid {
		rep.RespondedAt = respAt.Time.Format(time.RFC3339)
	} else {
		rep.RespondedAt = ""
	}

	rep.Attachments = []ReportAttachmentItem{}
	attRows, err := DB.Query(`
		SELECT id, mime_type, file_path, COALESCE(file_name, ''), COALESCE(file_size, 0)
		FROM evidence_files
		WHERE report_id = $1
		ORDER BY id ASC`, reportID)
	if err == nil {
		defer attRows.Close()
		for attRows.Next() {
			var att ReportAttachmentItem
			var encFilePath, encFileName string
			if err := attRows.Scan(&att.ID, &att.MimeType, &encFilePath, &encFileName, &att.FileSize); err == nil {
				if encFileName != "" {
					decName, err := DecryptAES(encFileName)
					if err == nil && decName != "" {
						att.FileName = decName
					}
				}
				if att.FileName == "" {
					decPath, err := DecryptAES(encFilePath)
					if err == nil {
						att.FileName = filepath.Base(decPath)
					} else {
						att.FileName = fmt.Sprintf("فایل پیوست شماره %d", len(rep.Attachments)+1)
					}
				}
				rep.Attachments = append(rep.Attachments, att)
			}
		}
	}
	rep.HasAttachment = len(rep.Attachments) > 0 || evidenceID.Valid

	msgRows, err := DB.Query(`
		SELECT rm.id, rm.sender_type, rm.message_text, COALESCE(rm.admin_username, ''), rm.created_at, COALESCE(NULLIF(au.full_name, ''), rm.admin_username, '')
		FROM report_messages rm
		LEFT JOIN admin_users au ON rm.admin_username = au.username
		WHERE rm.report_id = $1
		ORDER BY rm.created_at ASC, rm.id ASC`, reportID)
	rep.Messages = []ReportMessageItem{}
	if err == nil {
		defer msgRows.Close()
		for msgRows.Next() {
			var m ReportMessageItem
			var encMsg string
			var msgCreatedAt time.Time
			if err := msgRows.Scan(&m.ID, &m.SenderType, &encMsg, &m.AdminUsername, &msgCreatedAt, &m.AdminFullName); err == nil {
				decMsg, err := DecryptAES(encMsg)
				if err == nil {
					m.Message = decMsg
				} else {
					m.Message = "[پیام رمزنگاری‌شده]"
				}
				m.CreatedAt = msgCreatedAt.Format(time.RFC3339)
				rep.Messages = append(rep.Messages, m)
			}
		}
		if err := msgRows.Err(); err != nil {
			log.Printf("Error iterating report messages: %v", err)
		}
	}

	if admin.Role != "navahi_manategh" {
		rep.RegionalResponses = []RegionalResponseItem{}
		rRows, err := DB.Query(`
			SELECT rf.id, rf.recipient_username,
			       COALESCE(NULLIF(u.full_name, ''), rf.recipient_username) AS recip_name,
			       COALESCE(u.role_title, '') AS recip_title,
			       rf.response_text, rf.responded_at
			FROM report_forwardings rf
			LEFT JOIN admin_users u ON rf.recipient_username = u.username
			WHERE rf.report_id = $1 AND rf.response_text IS NOT NULL AND rf.response_text != ''
			ORDER BY rf.responded_at ASC, rf.id ASC
		`, reportID)
		if err == nil {
			defer rRows.Close()
			for rRows.Next() {
				var rr RegionalResponseItem
				var encResp sql.NullString
				var respAt sql.NullTime
				if err := rRows.Scan(&rr.ID, &rr.RecipientUsername, &rr.RecipientName, &rr.RecipientTitle, &encResp, &respAt); err == nil {
					if encResp.Valid && encResp.String != "" {
						dec, err := DecryptAES(encResp.String)
						if err == nil {
							rr.ResponseText = dec
						} else {
							rr.ResponseText = "[خطا در رمزگشایی پاسخ]"
						}
					}
					if respAt.Valid {
						rr.RespondedAt = respAt.Time.Format(time.RFC3339)
					}
					rep.RegionalResponses = append(rep.RegionalResponses, rr)
				}
			}
		}
	} else {
		var myRR RegionalResponseItem
		var encResp sql.NullString
		var respAt sql.NullTime
		err = DB.QueryRow(`
			SELECT rf.id, rf.recipient_username,
			       COALESCE(NULLIF(u.full_name, ''), rf.recipient_username) AS recip_name,
			       COALESCE(u.role_title, '') AS recip_title,
			       rf.response_text, rf.responded_at
			FROM report_forwardings rf
			LEFT JOIN admin_users u ON rf.recipient_username = u.username
			WHERE rf.report_id = $1 AND rf.recipient_username = $2
		`, reportID, admin.Username).Scan(&myRR.ID, &myRR.RecipientUsername, &myRR.RecipientName, &myRR.RecipientTitle, &encResp, &respAt)
		if err == nil {
			if encResp.Valid && encResp.String != "" {
				dec, err := DecryptAES(encResp.String)
				if err == nil {
					myRR.ResponseText = dec
				}
			}
			if respAt.Valid {
				myRR.RespondedAt = respAt.Time.Format(time.RFC3339)
			}
			rep.MyRegionalResponse = &myRR
		}
	}

	if admin.Role == "navahi_manategh" {
		rep.HasResponse = (rep.MyRegionalResponse != nil && strings.TrimSpace(rep.MyRegionalResponse.ResponseText) != "")
	} else {
		rep.HasResponse = (strings.TrimSpace(rep.AdminResponse) != "")
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(rep)
}

func UpdateInternalNotesHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	if admin.Role == "navahi_manategh" {
		http.Error(w, "شما به این قسمت دسترسی ندارید", http.StatusForbidden)
		return
	}

	var req struct {
		ReportID      int64  `json:"report_id"`
		InternalNotes string `json:"internal_notes"`
		Notes         string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	notesText := strings.TrimSpace(stripNul(req.InternalNotes))
	if notesText == "" {
		notesText = strings.TrimSpace(stripNul(req.Notes))
	}

	if utf8.RuneCountInString(notesText) > 4000 {
		http.Error(w, "یادداشت‌های محرمانه پرونده نمی‌تواند بیش از ۴۰۰۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, req.ReportID) {
		http.Error(w, "دسترسی غیرمجاز به این پرونده", http.StatusForbidden)
		return
	}

	var encNotes *string
	if notesText != "" {
		encrypted, err := EncryptAES(notesText)
		if err != nil {
			http.Error(w, "Encryption error", http.StatusInternalServerError)
			return
		}
		encNotes = &encrypted
	}

	_, err := DB.Exec("UPDATE reports SET internal_notes = $1, updated_at = NOW() WHERE id = $2", encNotes, req.ReportID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	logAudit(admin.Username, admin.Title, "UPDATE_INTERNAL_NOTES", &req.ReportID, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "یادداشت‌های محرمانه پرونده با موفقیت ذخیره و رمزنگاری شد."}`))
}

func SetReportPriorityHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	if admin.Role == "navahi_manategh" {
		http.Error(w, "شما به این قسمت دسترسی ندارید", http.StatusForbidden)
		return
	}

	var req struct {
		ReportID int64  `json:"report_id"`
		Priority string `json:"priority"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, req.ReportID) {
		http.Error(w, "دسترسی غیرمجاز به این پرونده", http.StatusForbidden)
		return
	}

	req.Priority = strings.TrimSpace(req.Priority)
	if req.Priority != "normal" && req.Priority != "urgent" && req.Priority != "critical" {
		http.Error(w, "Invalid priority level", http.StatusBadRequest)
		return
	}

	_, err := DB.Exec("UPDATE reports SET priority = $1, updated_at = NOW() WHERE id = $2", req.Priority, req.ReportID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	logAudit(admin.Username, admin.Title, "SET_PRIORITY_"+strings.ToUpper(req.Priority), &req.ReportID, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "اولویت گزارش با موفقیت بروزرسانی شد."}`))
}

func ClassifyReportHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	if admin.Role == "navahi_manategh" {
		http.Error(w, "دسترسی غیرمجاز: کارشناسان نواحی و مناطق مجاز به ارزیابی یا تغییر کیفیت گزارش نیستند.", http.StatusForbidden)
		return
	}

	var req struct {
		ReportID       int64  `json:"report_id"`
		Classification string `json:"classification"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "درخواست نامعتبر است.", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, req.ReportID) {
		http.Error(w, "دسترسی غیرمجاز به این پرونده", http.StatusForbidden)
		return
	}

	req.Classification = strings.TrimSpace(strings.ToLower(req.Classification))
	if req.Classification != "helpful" && req.Classification != "useless" {
		http.Error(w, "نوع ارزیابی نامعتبر است. گزینه‌های مجاز: helpful یا useless.", http.StatusBadRequest)
		return
	}

	if req.ReportID <= 0 {
		http.Error(w, "شناسه پرونده نامعتبر است.", http.StatusBadRequest)
		return
	}

	var existingClassification, existingClassifiedBy sql.NullString
	err := DB.QueryRow("SELECT classification, classified_by FROM reports WHERE id = $1", req.ReportID).
		Scan(&existingClassification, &existingClassifiedBy)
	if err == sql.ErrNoRows {
		http.Error(w, "پرونده گزارش مورد نظر یافت نشد.", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	isUpdate := existingClassification.Valid && existingClassification.String != ""
	if isUpdate {

		if !existingClassifiedBy.Valid || existingClassifiedBy.String != admin.Username {
			http.Error(w, "دسترسی غیرمجاز: تغییر ارزیابی تنها توسط کارشناس ثبت‌کننده اولیه امکان‌پذیر است.", http.StatusForbidden)
			return
		}
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`
		UPDATE reports
		SET classification = $1,
		    classified_by = $2,
		    classified_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3
	`, req.Classification, admin.Username, req.ReportID)
	if err != nil {
		http.Error(w, "خطا در ذخیره‌سازی ارزیابی گزارش", http.StatusInternalServerError)
		return
	}

	actionLog := "CLASSIFY_REPORT_" + strings.ToUpper(req.Classification)
	if isUpdate {
		actionLog = "RECLASSIFY_REPORT_" + strings.ToUpper(req.Classification)
	}

	if err := logAuditTx(tx, admin.Username, admin.Title, actionLog, &req.ReportID, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":             "success",
		"message":            "ارزیابی کیفیت پرونده با موفقیت ثبت گردید.",
		"classification":     req.Classification,
		"classified_by":      admin.Username,
		"classified_by_name": admin.FullName,
	})
}

func DownloadAttachmentHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	reportIDStr := r.URL.Query().Get("id")
	reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid report ID", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, reportID) {
		http.Error(w, "دسترسی غیرمجاز به فایل‌های ضمیمه این پرونده", http.StatusForbidden)
		return
	}

	fileIDStr := r.URL.Query().Get("file_id")

	absUploadDir, err := filepath.Abs("./secure_uploads")
	if err != nil {
		http.Error(w, "Internal configuration error", http.StatusInternalServerError)
		return
	}

	if fileIDStr != "" {
		fileID, _ := strconv.ParseInt(fileIDStr, 10, 64)
		var encPath, mimeType, encFileName string
		err = DB.QueryRow("SELECT file_path, mime_type, COALESCE(file_name, '') FROM evidence_files WHERE id = $1 AND report_id = $2", fileID, reportID).Scan(&encPath, &mimeType, &encFileName)
		if err == sql.ErrNoRows {
			http.Error(w, "No attachment found", http.StatusNotFound)
			return
		} else if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		realPath, err := DecryptAES(encPath)
		if err != nil {
			http.Error(w, "Failed to decrypt file path", http.StatusInternalServerError)
			return
		}

		cleanPath := filepath.Clean(realPath)
		absFilePath, err := filepath.Abs(cleanPath)
		if err != nil || !strings.HasPrefix(absFilePath, absUploadDir) {
			logAudit(admin.Username, admin.Title, "ILLEGAL_FILE_ACCESS_ATTEMPT", &reportID, getIP(r))
			http.Error(w, "Access Denied: Invalid file path boundary", http.StatusForbidden)
			return
		}

		encryptedBytes, err := os.ReadFile(absFilePath)
		if err != nil {
			http.Error(w, "Failed to read attachment from storage", http.StatusInternalServerError)
			return
		}

		decryptedBytes, err := DecryptFileBytes(encryptedBytes)
		if err != nil {
			http.Error(w, "Failed to decrypt attachment payload", http.StatusInternalServerError)
			return
		}

		logAudit(admin.Username, admin.Title, "DOWNLOAD_ATTACHMENT", &reportID, getIP(r))

		ext := filepath.Ext(cleanPath)
		downloadName := "evidence_" + reportIDStr + "_" + fileIDStr + ext
		if encFileName != "" {
			if decName, err := DecryptAES(encFileName); err == nil && decName != "" {
				downloadName = decName
			}
		}

		w.Header().Set("Content-Type", mimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", downloadName, url.PathEscape(downloadName)))
		w.Header().Set("Content-Length", strconv.Itoa(len(decryptedBytes)))
		w.WriteHeader(http.StatusOK)
		w.Write(decryptedBytes)
		return
	}

	type fileRow struct {
		id          int64
		encPath     string
		mimeType    string
		encFileName string
	}

	rows, err := DB.Query("SELECT id, file_path, mime_type, COALESCE(file_name, '') FROM evidence_files WHERE report_id = $1 ORDER BY id ASC", reportID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var files []fileRow
	for rows.Next() {
		var f fileRow
		if err := rows.Scan(&f.id, &f.encPath, &f.mimeType, &f.encFileName); err == nil {
			files = append(files, f)
		}
	}

	if len(files) == 0 {
		http.Error(w, "No attachment found", http.StatusNotFound)
		return
	}

	if len(files) == 1 {
		f := files[0]
		realPath, err := DecryptAES(f.encPath)
		if err != nil {
			http.Error(w, "Failed to decrypt file path", http.StatusInternalServerError)
			return
		}

		cleanPath := filepath.Clean(realPath)
		absFilePath, err := filepath.Abs(cleanPath)
		if err != nil || !strings.HasPrefix(absFilePath, absUploadDir) {
			logAudit(admin.Username, admin.Title, "ILLEGAL_FILE_ACCESS_ATTEMPT", &reportID, getIP(r))
			http.Error(w, "Access Denied: Invalid file path boundary", http.StatusForbidden)
			return
		}

		encryptedBytes, err := os.ReadFile(absFilePath)
		if err != nil {
			http.Error(w, "Failed to read attachment from storage", http.StatusInternalServerError)
			return
		}

		decryptedBytes, err := DecryptFileBytes(encryptedBytes)
		if err != nil {
			http.Error(w, "Failed to decrypt attachment payload", http.StatusInternalServerError)
			return
		}

		logAudit(admin.Username, admin.Title, "DOWNLOAD_ATTACHMENT", &reportID, getIP(r))

		ext := filepath.Ext(cleanPath)
		downloadName := "evidence_" + reportIDStr + ext
		if f.encFileName != "" {
			if decName, err := DecryptAES(f.encFileName); err == nil && decName != "" {
				downloadName = decName
			}
		}

		w.Header().Set("Content-Type", f.mimeType)
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", downloadName, url.PathEscape(downloadName)))
		w.Header().Set("Content-Length", strconv.Itoa(len(decryptedBytes)))
		w.WriteHeader(http.StatusOK)
		w.Write(decryptedBytes)
		return
	}

	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)
	usedNames := make(map[string]int)

	for _, f := range files {
		realPath, err := DecryptAES(f.encPath)
		if err != nil {
			continue
		}
		cleanPath := filepath.Clean(realPath)
		absFilePath, err := filepath.Abs(cleanPath)
		if err != nil || !strings.HasPrefix(absFilePath, absUploadDir) {
			continue
		}

		encryptedBytes, err := os.ReadFile(absFilePath)
		if err != nil {
			continue
		}

		decryptedBytes, err := DecryptFileBytes(encryptedBytes)
		if err != nil {
			continue
		}

		entryName := fmt.Sprintf("evidence_%d_%d%s", reportID, f.id, filepath.Ext(cleanPath))
		if f.encFileName != "" {
			if decName, err := DecryptAES(f.encFileName); err == nil && decName != "" {
				entryName = decName
			}
		}

		if count := usedNames[entryName]; count > 0 {
			ext := filepath.Ext(entryName)
			base := strings.TrimSuffix(entryName, ext)
			usedNames[entryName] = count + 1
			entryName = fmt.Sprintf("%s (%d)%s", base, count+1, ext)
		} else {
			usedNames[entryName] = 1
		}

		header := &zip.FileHeader{
			Name:     entryName,
			Method:   zip.Deflate,
			Modified: time.Now(),
		}
		header.SetMode(0600)

		entryWriter, err := zw.CreateHeader(header)
		if err != nil {
			continue
		}
		_, _ = entryWriter.Write(decryptedBytes)
	}

	if err := zw.Close(); err != nil {
		http.Error(w, "Failed to create archive", http.StatusInternalServerError)
		return
	}

	logAudit(admin.Username, admin.Title, "DOWNLOAD_ALL_ATTACHMENTS_ZIP", &reportID, getIP(r))

	zipDownloadName := fmt.Sprintf("پیوست‌های_گزارش_%s.zip", reportIDStr)
	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q; filename*=UTF-8''%s", zipDownloadName, url.PathEscape(zipDownloadName)))
	w.Header().Set("Content-Length", strconv.Itoa(buf.Len()))
	w.WriteHeader(http.StatusOK)
	w.Write(buf.Bytes())
}

func RespondToReportHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	if admin.Role == "navahi_manategh" {
		http.Error(w, "شما به این قسمت دسترسی ندارید", http.StatusForbidden)
		return
	}

	var req struct {
		ReportID int64  `json:"report_id"`
		Response string `json:"response"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.Response = strings.TrimSpace(stripNul(req.Response))
	if req.ReportID <= 0 || req.Response == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(req.Response) > 3000 {
		http.Error(w, "متن پاسخ نمی‌تواند بیش از ۳۰۰۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, req.ReportID) {
		http.Error(w, "دسترسی غیرمجاز به این پرونده", http.StatusForbidden)
		return
	}

	encResponse, err := EncryptAES(req.Response)
	if err != nil {
		http.Error(w, "Encryption failure", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(`
		UPDATE reports
		SET admin_response = $1, responded_at = NOW(), responded_by = $2, is_read = true, read_at = COALESCE(read_at, NOW()), updated_at = NOW(), has_new_message = false
		WHERE id = $3`,
		encResponse, admin.Username, req.ReportID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = DB.Exec(`
		INSERT INTO report_messages (report_id, sender_type, message_text, admin_username, created_at)
		VALUES ($1, 'admin', $2, $3, NOW())`,
		req.ReportID, encResponse, admin.Username)
	if err != nil {
		log.Printf("Failed to insert message into report_messages: %v", err)
	}

	GlobalSSEBroker.BroadcastNewMessage(req.ReportID)

	logAudit(admin.Username, admin.Title, "RESPOND_REPORT", &req.ReportID, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "پاسخ با موفقیت ثبت شد"}`))
}

type AdminUserItem struct {
	ID                 int64  `json:"id"`
	Username           string `json:"username"`
	FullName           string `json:"full_name"`
	RoleTitle          string `json:"role_title"`
	Role               string `json:"role"`
	IsActive           bool   `json:"is_active"`
	IsSuperuser        bool   `json:"is_superuser"`
	PasswordChangedAt  string `json:"password_changed_at"`
	MustChangePassword bool   `json:"must_change_password"`
	IsPasswordExpired  bool   `json:"is_password_expired"`
	IsLocked           bool   `json:"is_locked"`
	LockoutRemaining   string `json:"lockout_remaining,omitempty"`
	CreatedAt          string `json:"created_at"`
	TotpEnabled        bool   `json:"totp_enabled"`
}

func GetUsersHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(`
		SELECT id, username, COALESCE(full_name, ''), role_title, COALESCE(role, 'edare_kol'), is_active, is_superuser,
		       COALESCE(password_changed_at, NOW()), COALESCE(must_change_password, false), created_at,
		       COALESCE(totp_enabled, false)
		FROM admin_users
		ORDER BY id ASC
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var users []AdminUserItem
	for rows.Next() {
		var u AdminUserItem
		var pwdChangedAt, createdAt time.Time
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.RoleTitle, &u.Role, &u.IsActive, &u.IsSuperuser, &pwdChangedAt, &u.MustChangePassword, &createdAt, &u.TotpEnabled); err != nil {
			continue
		}
		u.PasswordChangedAt = pwdChangedAt.Format(time.RFC3339)
		u.IsPasswordExpired = time.Since(pwdChangedAt) > 30*24*time.Hour
		u.CreatedAt = createdAt.Format(time.RFC3339)

		if isLocked, remaining := Lockouts.IsLocked(u.Username); isLocked {
			u.IsLocked = true
			u.LockoutRemaining = fmt.Sprintf("%d دقیقه", int(remaining.Minutes())+1)
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if users == nil {
		users = []AdminUserItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		Username           string `json:"username"`
		Password           string `json:"password"`
		FullName           string `json:"full_name"`
		RoleTitle          string `json:"role_title"`
		Role               string `json:"role"`
		MustChangePassword bool   `json:"must_change_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Password = strings.TrimSpace(req.Password)
	req.FullName = strings.TrimSpace(req.FullName)
	req.RoleTitle = strings.TrimSpace(req.RoleTitle)
	req.Role = strings.TrimSpace(req.Role)

	if req.Username == "" || req.FullName == "" || req.RoleTitle == "" {
		http.Error(w, "اطلاعات ورودی ناقص است (نام و نام خانوادگی، نام کاربری و سمت سازمانی الزامی است).", http.StatusBadRequest)
		return
	}

	if req.Role != "superuser" && req.Role != "navahi_manategh" {
		req.Role = "edare_kol"
	}
	isSuper := (req.Role == "superuser")

	if err := ValidatePasswordComplexity(req.Password); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	passHash, err := HashPassword(req.Password)
	if err != nil {
		http.Error(w, "خطا در پردازش امن رمز عبور", http.StatusInternalServerError)
		return
	}

	var newID int64

	err = DB.QueryRow(`
		INSERT INTO admin_users (username, password_hash, full_name, role_title, role, is_active, is_superuser, password_changed_at, must_change_password)
		VALUES ($1, $2, $3, $4, $5, true, $6, NOW(), $7) RETURNING id`,
		req.Username, passHash, req.FullName, req.RoleTitle, req.Role, isSuper, req.MustChangePassword).Scan(&newID)
	if err != nil {
		if strings.Contains(err.Error(), "unique") || strings.Contains(err.Error(), "duplicate") {
			http.Error(w, "این نام کاربری قبلاً در سامانه ثبت شده است.", http.StatusConflict)
			return
		}
		http.Error(w, "خطا در ثبت کاربر در دیتابیس", http.StatusInternalServerError)
		return
	}

	_, _ = DB.Exec("INSERT INTO admin_password_history (admin_username, user_id, password_hash, created_at) VALUES ($1, $2, $3, NOW())", req.Username, newID, passHash)

	logAudit(admin.Username, admin.Title, "CREATE_ADMIN_USER:"+req.Username, nil, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "کاربر با موفقیت ایجاد شد",
		"user_id": newID,
	})
}

func ToggleUserActiveHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var targetUsername string
	var isSuper bool
	err := DB.QueryRow("SELECT username, is_superuser FROM admin_users WHERE id = $1", req.UserID).Scan(&targetUsername, &isSuper)
	if err != nil {
		http.Error(w, "کاربر یافت نشد", http.StatusNotFound)
		return
	}

	if isSuper {
		http.Error(w, "امکان تغییر وضعیت این حساب کاربری توسط پشتیبانی سامانه وجود ندارد.", http.StatusBadRequest)
		return
	}

	if targetUsername == admin.Username {
		http.Error(w, "نمی‌توانید وضعیت حساب کاربری خودتان را غیرفعال کنید.", http.StatusBadRequest)
		return
	}

	var isActive bool
	_ = DB.QueryRow("UPDATE admin_users SET is_active = NOT is_active WHERE id = $1 RETURNING is_active", req.UserID).Scan(&isActive)
	if !isActive {
		Sessions.RevokeAllUserSessions(targetUsername, "DEACTIVATED")
	}

	logAudit(admin.Username, admin.Title, "TOGGLE_USER_ACTIVE:"+targetUsername, nil, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "وضعیت کاربر با موفقیت تغییر یافت"}`))
}

func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		UserID             int64  `json:"user_id"`
		FullName           string `json:"full_name"`
		RoleTitle          string `json:"role_title"`
		Role               string `json:"role"`
		IsActive           bool   `json:"is_active"`
		MustChangePassword bool   `json:"must_change_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.FullName = strings.TrimSpace(req.FullName)
	req.RoleTitle = strings.TrimSpace(req.RoleTitle)
	req.Role = strings.TrimSpace(req.Role)

	if req.FullName == "" || req.RoleTitle == "" {
		http.Error(w, "نام و نام خانوادگی و سمت سازمانی الزامی است.", http.StatusBadRequest)
		return
	}

	if req.Role != "superuser" && req.Role != "navahi_manategh" {
		req.Role = "edare_kol"
	}

	var targetUsername, currentRole string
	var isSuper bool
	err := DB.QueryRow("SELECT username, is_superuser, COALESCE(role, 'edare_kol') FROM admin_users WHERE id = $1", req.UserID).Scan(&targetUsername, &isSuper, &currentRole)
	if err != nil {
		http.Error(w, "کاربر یافت نشد", http.StatusNotFound)
		return
	}

	if (isSuper || currentRole == "superuser") && req.Role != "superuser" {
		var otherSuperCount int
		_ = DB.QueryRow("SELECT COUNT(*) FROM admin_users WHERE (is_superuser = true OR role = 'superuser') AND is_active = true AND id != $1", req.UserID).Scan(&otherSuperCount)
		if otherSuperCount < 1 {
			http.Error(w, "امکان تغییر نقش این کاربر وجود ندارد: حداقل یک مدیر ارشد سامانه فعال باید در سیستم باقی بماند.", http.StatusBadRequest)
			return
		}
	}

	if targetUsername == admin.Username && !req.IsActive {
		http.Error(w, "نمی‌توانید حساب کاربری خودتان را غیرفعال کنید.", http.StatusBadRequest)
		return
	}

	isNewSuper := (req.Role == "superuser")

	_, err = DB.Exec(`
		UPDATE admin_users
		SET full_name = $1, role_title = $2, role = $3, is_superuser = $4, is_active = $5, must_change_password = $6
		WHERE id = $7
	`, req.FullName, req.RoleTitle, req.Role, isNewSuper, req.IsActive, req.MustChangePassword, req.UserID)

	if err != nil {
		http.Error(w, "خطای دیتابیس در ویرایش کاربر", http.StatusInternalServerError)
		return
	}

	if !req.IsActive || currentRole != req.Role {
		Sessions.RevokeAllUserSessions(targetUsername, "ROLE_OR_STATUS_CHANGED")
	}

	logAudit(admin.Username, admin.Title, "UPDATE_ADMIN_USER:"+targetUsername, nil, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "اطلاعات کاربر با موفقیت بروزرسانی شد."}`))
}

func DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	var targetUsername string
	var isSuper bool
	err := DB.QueryRow("SELECT username, is_superuser FROM admin_users WHERE id = $1", req.UserID).Scan(&targetUsername, &isSuper)
	if err != nil {
		http.Error(w, "کاربر یافت نشد", http.StatusNotFound)
		return
	}

	if isSuper {
		http.Error(w, "حساب کاربری اصلی پشتیبانی سامانه غیرقابل حذف است.", http.StatusBadRequest)
		return
	}

	if targetUsername == admin.Username {
		http.Error(w, "نمی‌توانید حساب کاربری خودتان را حذف نمایید.", http.StatusBadRequest)
		return
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM admin_users WHERE id = $1", req.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := logAuditTx(tx, admin.Username, admin.Title, "DELETE_ADMIN_USER:"+targetUsername, nil, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	Sessions.RevokeAllUserSessions(targetUsername, "DELETED")

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "کاربر با موفقیت حذف شد"}`))
}

func ResetUserPasswordHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		UserID             int64  `json:"user_id"`
		NewPassword        string `json:"new_password"`
		MustChangePassword bool   `json:"must_change_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.UserID <= 0 {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.NewPassword = strings.TrimSpace(req.NewPassword)
	if err := ValidatePasswordComplexity(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var targetUsername string
	err := DB.QueryRow("SELECT username FROM admin_users WHERE id = $1", req.UserID).Scan(&targetUsername)
	if err != nil {
		http.Error(w, "کاربر یافت نشد", http.StatusNotFound)
		return
	}

	passHash, err := HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "خطا در رمزنگاری پسورد", http.StatusInternalServerError)
		return
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE admin_users SET password_hash = $1, password_changed_at = NOW(), must_change_password = $2 WHERE id = $3", passHash, req.MustChangePassword, req.UserID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO admin_password_history (admin_username, user_id, password_hash, created_at) VALUES ($1, $2, $3, NOW())", targetUsername, req.UserID, passHash)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if err := logAuditTx(tx, admin.Username, admin.Title, "RESET_PASSWORD_USER:"+targetUsername, nil, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "رمز عبور با موفقیت تغییر کرد"}`))
}

func LiftUserLockoutHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" && req.UserID > 0 {
		_ = DB.QueryRow("SELECT username FROM admin_users WHERE id = $1", req.UserID).Scan(&req.Username)
	}
	if req.Username == "" {
		http.Error(w, "نام کاربری مشخص نشده است.", http.StatusBadRequest)
		return
	}

	Lockouts.Reset(req.Username)
	logAudit(admin.Username, admin.Title, "LIFT_LOCKOUT:"+req.Username, nil, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "قفل موقت ورود با موفقیت برداشته شد و کاربر می‌تواند وارد شود."}`))
}

func DisableUser2FAHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		UserID   int64  `json:"user_id"`
		Username string `json:"username"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" && req.UserID > 0 {
		_ = DB.QueryRow("SELECT username FROM admin_users WHERE id = $1", req.UserID).Scan(&req.Username)
	}
	if req.Username == "" {
		http.Error(w, "نام کاربری یا شناسه کاربر مشخص نشده است.", http.StatusBadRequest)
		return
	}

	var targetRoleTitle string
	var targetID int64
	err := DB.QueryRow("SELECT id, role_title FROM admin_users WHERE username = $1", req.Username).Scan(&targetID, &targetRoleTitle)
	if err != nil {
		http.Error(w, "کاربر مورد نظر یافت نشد.", http.StatusNotFound)
		return
	}

	_, err = DB.Exec("UPDATE admin_users SET totp_secret = NULL, totp_enabled = false WHERE id = $1", targetID)
	if err != nil {
		http.Error(w, "خطای پایگاه داده در غیرفعال‌سازی احراز هویت دومرحله‌ای", http.StatusInternalServerError)
		return
	}

	logAudit(admin.Username, admin.Title, "SUPERUSER_DISABLE_2FA:"+req.Username, &targetID, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "احراز هویت دومرحله‌ای برای این حساب با موفقیت غیرفعال شد."}`))
}

type AuditLogItem struct {
	ID                 int64   `json:"id"`
	AdminUsername      string  `json:"admin_username"`
	AdminTitle         string  `json:"admin_title"`
	Action             string  `json:"action"`
	TargetReportID     *int64  `json:"target_report_id"`
	TargetTrackingCode *string `json:"target_tracking_code"`
	IPAddress          string  `json:"ip_address"`
	CreatedAt          string  `json:"created_at"`
}

func GetAuditLogsHandler(w http.ResponseWriter, r *http.Request) {
	rows, err := DB.Query(`
		SELECT a.id, a.admin_username, a.admin_title, a.action, a.target_report_id, r.tracking_code, a.ip_address, a.created_at
		FROM audit_logs a
		LEFT JOIN reports r ON a.target_report_id = r.id
		ORDER BY a.id DESC
		LIMIT 150
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var logs []AuditLogItem
	for rows.Next() {
		var l AuditLogItem
		var targetID sql.NullInt64
		var trackingCode sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&l.ID, &l.AdminUsername, &l.AdminTitle, &l.Action, &targetID, &trackingCode, &l.IPAddress, &createdAt); err != nil {
			continue
		}
		if targetID.Valid {
			val := targetID.Int64
			l.TargetReportID = &val
		}
		if trackingCode.Valid && trackingCode.String != "" {
			tCode := trackingCode.String
			l.TargetTrackingCode = &tCode
		}
		l.CreatedAt = createdAt.Format(time.RFC3339)
		logs = append(logs, l)
	}

	if err := rows.Err(); err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	if logs == nil {
		logs = []AuditLogItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(logs)
}

func GetSystemStatusHandler(w http.ResponseWriter, r *http.Request) {

	dbStatus := "متصل و فعال (Healthy)"
	if err := DB.Ping(); err != nil {
		dbStatus = "خطا در اتصال: " + err.Error()
	}

	var dbNow time.Time
	var dbTimezone string
	_ = DB.QueryRow("SELECT NOW(), CURRENT_SETTING('TIMEZONE')").Scan(&dbNow, &dbTimezone)

	var totalReports int64
	var unreadReports int64
	var totalUsers int64
	var totalAuditLogs int64

	_ = DB.QueryRow("SELECT COUNT(*) FROM reports").Scan(&totalReports)
	_ = DB.QueryRow("SELECT COUNT(*) FROM reports WHERE is_read = false").Scan(&unreadReports)
	_ = DB.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&totalUsers)
	_ = DB.QueryRow("SELECT COUNT(*) FROM audit_logs").Scan(&totalAuditLogs)

	stats := DB.Stats()

	resp := map[string]interface{}{
		"db_status":        dbStatus,
		"db_timezone":      dbTimezone,
		"db_time":          dbNow.Format(time.RFC3339),
		"db_open_conns":    stats.OpenConnections,
		"db_idle_conns":    stats.Idle,
		"db_in_use_conns":  stats.InUse,
		"active_sessions":  Sessions.ActiveCount(),
		"server_time":      time.Now().Format(time.RFC3339),
		"total_reports":    totalReports,
		"unread_reports":   unreadReports,
		"total_users":      totalUsers,
		"total_audit_logs": totalAuditLogs,
		"audit_failures":   atomic.LoadUint64(&auditFailures),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func GetServerTimeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	var dbNow time.Time
	var dbTz string
	var err error = fmt.Errorf("db unavailable")
	if DB != nil {
		err = DB.QueryRow("SELECT NOW(), CURRENT_SETTING('TIMEZONE')").Scan(&dbNow, &dbTz)
	}
	if err != nil {
		dbNow = time.Now()
		dbTz = "Asia/Tehran"
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("X-Server-Time", dbNow.Format(time.RFC3339Nano))

	json.NewEncoder(w).Encode(map[string]interface{}{
		"server_time":         dbNow.Format(time.RFC3339Nano),
		"server_time_unix_ms": dbNow.UnixMilli(),
		"timezone":            dbTz,
	})
}

func GetActiveSessionsHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := r.Context().Value("admin").(AdminSession)
	if !ok || !admin.IsSuperuser {
		http.Error(w, "دسترسی غیرمجاز: دسترسی به این بخش محدود به پشتیبانی سامانه می‌باشد.", http.StatusForbidden)
		return
	}

	sessions := Sessions.GetAllActiveSessions()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}

func TerminateSessionHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	admin, ok := r.Context().Value("admin").(AdminSession)
	if !ok || !admin.IsSuperuser {
		http.Error(w, "دسترسی غیرمجاز: دسترسی به این بخش محدود به پشتیبانی سامانه می‌باشد.", http.StatusForbidden)
		return
	}

	var req struct {
		Token string `json:"token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Token == "" {
		http.Error(w, "شناسه نشست نامعتبر است.", http.StatusBadRequest)
		return
	}

	tokenHash := req.Token
	if len(req.Token) != 64 {
		tokenHash = hashToken(req.Token)
	} else {
		for _, c := range req.Token {
			if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
				tokenHash = hashToken(req.Token)
				break
			}
		}
	}

	var targetUsername string
	err := DB.QueryRow("SELECT username FROM admin_sessions WHERE token_hash = $1", tokenHash).Scan(&targetUsername)
	if err == sql.ErrNoRows {
		if s, ok := Sessions.Get(req.Token); ok {
			targetUsername = s.Username
		} else {
			http.Error(w, "نشست مورد نظر یافت نشد یا قبلاً منقضی شده است.", http.StatusNotFound)
			return
		}
	} else if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec(`INSERT INTO admin_session_kicks (token_hash, reason, created_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (token_hash) DO UPDATE SET reason = EXCLUDED.reason`, tokenHash, "TERMINATED_BY_SUPERUSER")
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec(`DELETE FROM admin_sessions WHERE token_hash = $1`, tokenHash)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	if err := logAuditTx(tx, admin.Username, admin.Title, "TERMINATE_SESSION_USER:"+targetUsername, nil, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	Sessions.mu.Lock()
	delete(Sessions.cache, tokenHash)
	delete(Sessions.sessions, req.Token)
	for t := range Sessions.sessions {
		if hashToken(t) == tokenHash {
			delete(Sessions.sessions, t)
			Sessions.kickedTokens[t] = "TERMINATED_BY_SUPERUSER"
		}
	}
	Sessions.kickedTokens[req.Token] = "TERMINATED_BY_SUPERUSER"
	Sessions.mu.Unlock()

	GlobalSSEBroker.BroadcastSessionRevoked(targetUsername, tokenHash)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "نشست کاربر با موفقیت خاتمه یافت."}`))
}

func ChangeMyPasswordHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	req.CurrentPassword = strings.TrimSpace(req.CurrentPassword)
	req.NewPassword = strings.TrimSpace(req.NewPassword)

	if req.NewPassword == "" || req.CurrentPassword == "" {
		http.Error(w, "اطلاعات ورودی ناقص است.", http.StatusBadRequest)
		return
	}

	if req.NewPassword == req.CurrentPassword {
		http.Error(w, "کلمه عبور جدید نمی‌تواند با کلمه عبور فعلی یکسان باشد.", http.StatusBadRequest)
		return
	}

	if err := ValidatePasswordComplexity(req.NewPassword); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var userID int64
	var currentHash string
	var passwordChangedAt time.Time
	var mustChangePassword bool
	err := DB.QueryRow("SELECT id, password_hash, password_changed_at, must_change_password FROM admin_users WHERE username = $1 AND is_active = true", admin.Username).Scan(&userID, &currentHash, &passwordChangedAt, &mustChangePassword)
	if err != nil {
		http.Error(w, "حساب کاربری یافت نشد یا غیرفعال است.", http.StatusUnauthorized)
		return
	}

	if !mustChangePassword && time.Since(passwordChangedAt) < 24*time.Hour {
		http.Error(w, "امکان تغییر کلمه عبور زودتر از ۲۴ ساعت از آخرین تغییر وجود ندارد.", http.StatusBadRequest)
		return
	}

	if !VerifyPassword(req.CurrentPassword, currentHash) {
		http.Error(w, "کلمه عبور فعلی وارد شده نادرست است.", http.StatusUnauthorized)
		return
	}

	if VerifyPassword(req.NewPassword, currentHash) {
		http.Error(w, "کلمه عبور جدید نمی‌تواند با کلمه عبور فعلی یکسان باشد.", http.StatusBadRequest)
		return
	}

	rows, err := DB.Query("SELECT password_hash FROM admin_password_history WHERE admin_username = $1 OR user_id = $2 ORDER BY created_at DESC LIMIT 5", admin.Username, userID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var oldHash string
			if err := rows.Scan(&oldHash); err == nil {
				if VerifyPassword(req.NewPassword, oldHash) {
					http.Error(w, "کلمه عبور جدید نمی‌تواند مشابه ۵ کلمه عبور اخیر شما باشد.", http.StatusBadRequest)
					return
				}
			}
		}
	}

	newHash, err := HashPassword(req.NewPassword)
	if err != nil {
		http.Error(w, "خطا در رمزنگاری کلمه عبور جدید", http.StatusInternalServerError)
		return
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE admin_users SET password_hash = $1, password_changed_at = NOW(), must_change_password = false WHERE id = $2", newHash, userID)
	if err != nil {
		http.Error(w, "خطای دیتابیس در ذخیره رمز جدید", http.StatusInternalServerError)
		return
	}

	_, err = tx.Exec("INSERT INTO admin_password_history (admin_username, user_id, password_hash, created_at) VALUES ($1, $2, $3, NOW())", admin.Username, userID, newHash)
	if err != nil {
		http.Error(w, "خطای دیتابیس در ثبت تاریخچه رمز عبور", http.StatusInternalServerError)
		return
	}

	if err := logAuditTx(tx, admin.Username, admin.Title, "CHANGE_OWN_PASSWORD", nil, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status": "success", "message": "کلمه عبور شما با موفقیت تغییر یافت."}`))
}

func DeleteReportHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	admin, ok := r.Context().Value("admin").(AdminSession)
	if !ok || !admin.IsSuperuser {
		logAudit("UNKNOWN", "SECURITY_ALERT", "UNAUTHORIZED_DELETE_REPORT_ATTEMPT", nil, getIP(r))
		http.Error(w, "دسترسی غیرمجاز: عملیات حذف پرونده منحصراً در اختیار پشتیبانی سامانه می‌باشد.", http.StatusForbidden)
		return
	}

	var req struct {
		ReportID  int64   `json:"report_id"`
		ReportIDs []int64 `json:"report_ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "داده‌های درخواست نامعتبر است", http.StatusBadRequest)
		return
	}

	idMap := make(map[int64]bool)
	if req.ReportID > 0 {
		idMap[req.ReportID] = true
	}
	for _, id := range req.ReportIDs {
		if id > 0 {
			idMap[id] = true
		}
	}

	var ids []int64
	for id := range idMap {
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		http.Error(w, "شناسه پرونده نامعتبر است", http.StatusBadRequest)
		return
	}

	var filesToDelete []string
	rows, err := DB.Query("SELECT file_path FROM evidence_files WHERE report_id = ANY($1)", pq.Array(ids))
	if err == nil {
		defer rows.Close()
		absUploadDir, _ := filepath.Abs("./secure_uploads")

		for rows.Next() {
			var encPath string
			if err := rows.Scan(&encPath); err == nil {
				if realPath, err := DecryptAES(encPath); err == nil {
					cleanPath := filepath.Clean(realPath)
					absFilePath, err := filepath.Abs(cleanPath)
					if err == nil && strings.HasPrefix(absFilePath, absUploadDir) {
						filesToDelete = append(filesToDelete, absFilePath)
					}
				}
			}
		}
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM reports WHERE id = ANY($1)", pq.Array(ids))
	if err != nil {
		http.Error(w, "خطا در حذف رکورد پرونده از پایگاه داده", http.StatusInternalServerError)
		return
	}

	for _, repID := range ids {
		repIDCopy := repID
		if err := logAuditTx(tx, admin.Username, admin.Title, "DELETE_REPORT_AND_ATTACHMENTS", &repIDCopy, getIP(r)); err != nil {
			http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
			return
		}
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "خطای پایگاه داده", http.StatusInternalServerError)
		return
	}

	for _, absFilePath := range filesToDelete {
		if err := os.Remove(absFilePath); err != nil && !os.IsNotExist(err) {
			_ = logAudit(admin.Username, admin.Title, "FILE_DELETE_WARNING", nil, getIP(r))
		}
	}

	var msg string
	if len(ids) == 1 {
		msg = "پرونده و فایل‌های ضمیمه آن با موفقیت حذف گردیدند."
	} else {
		msg = fmt.Sprintf("%d پرونده با موفقیت حذف گردیدند.", len(ids))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":        "success",
		"message":       msg,
		"deleted_count": len(ids),
	})
}

type AnalyticsResponse struct {
	CategoryBreakdown       []CategoryCount         `json:"category_breakdown"`
	MonthlyTrend            []MonthCount            `json:"monthly_trend"`
	PriorityBreakdown       []PriorityCount         `json:"priority_breakdown"`
	TypeBreakdown           []TypeCount             `json:"type_breakdown"`
	AvgResponseHours        float64                 `json:"avg_response_hours"`
	TotalReports            int64                   `json:"total_reports"`
	TotalResponded          int64                   `json:"total_responded"`
	TotalUnread             int64                   `json:"total_unread"`
	ClassificationAnalytics ClassificationAnalytics `json:"classification_analytics"`
}

type ClassificationCount struct {
	Classification string  `json:"classification"`
	Count          int64   `json:"count"`
	Percentage     float64 `json:"percentage"`
}

type ClassificationAnalytics struct {
	TotalClassified   int64                 `json:"total_classified"`
	TotalHelpful      int64                 `json:"total_helpful"`
	TotalUseless      int64                 `json:"total_useless"`
	HelpfulPercentage float64               `json:"helpful_percentage"`
	UselessPercentage float64               `json:"useless_percentage"`
	Breakdown         []ClassificationCount `json:"breakdown"`
}

type CategoryCount struct {
	Category string `json:"category"`
	Count    int64  `json:"count"`
}

type MonthCount struct {
	Month string `json:"month"`
	Count int64  `json:"count"`
}

type PriorityCount struct {
	Priority string `json:"priority"`
	Count    int64  `json:"count"`
}

type TypeCount struct {
	ReportType string `json:"report_type"`
	Count      int64  `json:"count"`
}

func GetAnalyticsHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)
	var resp AnalyticsResponse

	if admin.Role == "navahi_manategh" {
		catRows, err := DB.Query(`
			SELECT r.category, COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1
			GROUP BY r.category
			ORDER BY COUNT(DISTINCT r.id) DESC
		`, admin.Username)
		if err == nil {
			defer catRows.Close()
			for catRows.Next() {
				var c CategoryCount
				if err := catRows.Scan(&c.Category, &c.Count); err == nil {
					resp.CategoryBreakdown = append(resp.CategoryBreakdown, c)
				}
			}
		}

		monthRows, err := DB.Query(`
			SELECT TO_CHAR(r.created_at, 'YYYY-MM') AS month, COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1 AND r.created_at >= NOW() - INTERVAL '12 months'
			GROUP BY month
			ORDER BY month ASC
		`, admin.Username)
		if err == nil {
			defer monthRows.Close()
			for monthRows.Next() {
				var m MonthCount
				if err := monthRows.Scan(&m.Month, &m.Count); err == nil {
					resp.MonthlyTrend = append(resp.MonthlyTrend, m)
				}
			}
		}

		prioRows, err := DB.Query(`
			SELECT COALESCE(r.priority, 'normal'), COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1
			GROUP BY r.priority
			ORDER BY COUNT(DISTINCT r.id) DESC
		`, admin.Username)
		if err == nil {
			defer prioRows.Close()
			for prioRows.Next() {
				var p PriorityCount
				if err := prioRows.Scan(&p.Priority, &p.Count); err == nil {
					resp.PriorityBreakdown = append(resp.PriorityBreakdown, p)
				}
			}
		}

		typeRows, err := DB.Query(`
			SELECT r.report_type, COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1
			GROUP BY r.report_type
			ORDER BY COUNT(DISTINCT r.id) DESC
		`, admin.Username)
		if err == nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var t TypeCount
				if err := typeRows.Scan(&t.ReportType, &t.Count); err == nil {
					resp.TypeBreakdown = append(resp.TypeBreakdown, t)
				}
			}
		}

		_ = DB.QueryRow(`
			SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (rf.responded_at - rf.forwarded_at)) / 3600.0), 0)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1 AND rf.responded_at IS NOT NULL
		`, admin.Username).Scan(&resp.AvgResponseHours)

		_ = DB.QueryRow(`
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1
		`, admin.Username).Scan(&resp.TotalReports)

		_ = DB.QueryRow(`
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1 AND rf.response_text IS NOT NULL AND rf.response_text != ''
		`, admin.Username).Scan(&resp.TotalResponded)

		_ = DB.QueryRow(`
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			LEFT JOIN admin_report_views arv ON r.id = arv.report_id AND arv.admin_username = $1
			WHERE rf.recipient_username = $1 AND arv.id IS NULL
		`, admin.Username).Scan(&resp.TotalUnread)

		var totalHelpful, totalUseless int64
		_ = DB.QueryRow(`
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1 AND r.classification = 'helpful'
		`, admin.Username).Scan(&totalHelpful)

		_ = DB.QueryRow(`
			SELECT COUNT(DISTINCT r.id)
			FROM reports r
			INNER JOIN report_forwardings rf ON r.id = rf.report_id
			WHERE rf.recipient_username = $1 AND r.classification = 'useless'
		`, admin.Username).Scan(&totalUseless)

		totalClassified := totalHelpful + totalUseless
		var helpfulPct, uselessPct float64
		if totalClassified > 0 {
			helpfulPct = math.Round((float64(totalHelpful)/float64(totalClassified))*1000) / 10.0
			uselessPct = math.Round((float64(totalUseless)/float64(totalClassified))*1000) / 10.0
		}

		resp.ClassificationAnalytics = ClassificationAnalytics{
			TotalClassified:   totalClassified,
			TotalHelpful:      totalHelpful,
			TotalUseless:      totalUseless,
			HelpfulPercentage: helpfulPct,
			UselessPercentage: uselessPct,
			Breakdown: []ClassificationCount{
				{Classification: "helpful", Count: totalHelpful, Percentage: helpfulPct},
				{Classification: "useless", Count: totalUseless, Percentage: uselessPct},
			},
		}
	} else {
		catRows, err := DB.Query("SELECT category, COUNT(*) FROM reports GROUP BY category ORDER BY COUNT(*) DESC")
		if err == nil {
			defer catRows.Close()
			for catRows.Next() {
				var c CategoryCount
				if err := catRows.Scan(&c.Category, &c.Count); err == nil {
					resp.CategoryBreakdown = append(resp.CategoryBreakdown, c)
				}
			}
			if err := catRows.Err(); err != nil {
				log.Printf("Error iterating category breakdown: %v", err)
			}
		}

		monthRows, err := DB.Query(`
			SELECT TO_CHAR(created_at, 'YYYY-MM') AS month, COUNT(*)
			FROM reports
			WHERE created_at >= NOW() - INTERVAL '12 months'
			GROUP BY month
			ORDER BY month ASC
		`)
		if err == nil {
			defer monthRows.Close()
			for monthRows.Next() {
				var m MonthCount
				if err := monthRows.Scan(&m.Month, &m.Count); err == nil {
					resp.MonthlyTrend = append(resp.MonthlyTrend, m)
				}
			}
			if err := monthRows.Err(); err != nil {
				log.Printf("Error iterating monthly trend: %v", err)
			}
		}

		prioRows, err := DB.Query("SELECT COALESCE(priority, 'normal'), COUNT(*) FROM reports GROUP BY priority ORDER BY COUNT(*) DESC")
		if err == nil {
			defer prioRows.Close()
			for prioRows.Next() {
				var p PriorityCount
				if err := prioRows.Scan(&p.Priority, &p.Count); err == nil {
					resp.PriorityBreakdown = append(resp.PriorityBreakdown, p)
				}
			}
			if err := prioRows.Err(); err != nil {
				log.Printf("Error iterating priority breakdown: %v", err)
			}
		}

		typeRows, err := DB.Query("SELECT report_type, COUNT(*) FROM reports GROUP BY report_type ORDER BY COUNT(*) DESC")
		if err == nil {
			defer typeRows.Close()
			for typeRows.Next() {
				var t TypeCount
				if err := typeRows.Scan(&t.ReportType, &t.Count); err == nil {
					resp.TypeBreakdown = append(resp.TypeBreakdown, t)
				}
			}
			if err := typeRows.Err(); err != nil {
				log.Printf("Error iterating type breakdown: %v", err)
			}
		}

		_ = DB.QueryRow(`
			SELECT COALESCE(AVG(EXTRACT(EPOCH FROM (responded_at - created_at)) / 3600.0), 0)
			FROM reports
			WHERE responded_at IS NOT NULL
		`).Scan(&resp.AvgResponseHours)

		_ = DB.QueryRow("SELECT COUNT(*) FROM reports").Scan(&resp.TotalReports)
		_ = DB.QueryRow("SELECT COUNT(*) FROM reports WHERE responded_at IS NOT NULL").Scan(&resp.TotalResponded)
		_ = DB.QueryRow("SELECT COUNT(*) FROM reports WHERE is_read = false").Scan(&resp.TotalUnread)

		var totalHelpful, totalUseless int64
		_ = DB.QueryRow("SELECT COUNT(*) FROM reports WHERE classification = 'helpful'").Scan(&totalHelpful)
		_ = DB.QueryRow("SELECT COUNT(*) FROM reports WHERE classification = 'useless'").Scan(&totalUseless)
		totalClassified := totalHelpful + totalUseless

		var helpfulPct, uselessPct float64
		if totalClassified > 0 {
			helpfulPct = math.Round((float64(totalHelpful)/float64(totalClassified))*1000) / 10.0
			uselessPct = math.Round((float64(totalUseless)/float64(totalClassified))*1000) / 10.0
		}

		resp.ClassificationAnalytics = ClassificationAnalytics{
			TotalClassified:   totalClassified,
			TotalHelpful:      totalHelpful,
			TotalUseless:      totalUseless,
			HelpfulPercentage: helpfulPct,
			UselessPercentage: uselessPct,
			Breakdown: []ClassificationCount{
				{Classification: "helpful", Count: totalHelpful, Percentage: helpfulPct},
				{Classification: "useless", Count: totalUseless, Percentage: uselessPct},
			},
		}
	}

	if resp.CategoryBreakdown == nil {
		resp.CategoryBreakdown = []CategoryCount{}
	}
	if resp.MonthlyTrend == nil {
		resp.MonthlyTrend = []MonthCount{}
	}
	if resp.PriorityBreakdown == nil {
		resp.PriorityBreakdown = []PriorityCount{}
	}
	if resp.TypeBreakdown == nil {
		resp.TypeBreakdown = []TypeCount{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

type RecipientUserItem struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	RoleTitle string `json:"role_title"`
}

func GetForwardingRecipientsHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)
	if admin.Role != "superuser" && admin.Role != "edare_kol" && !admin.IsSuperuser {
		http.Error(w, "دسترسی غیرمجاز", http.StatusForbidden)
		return
	}

	rows, err := DB.Query(`
		SELECT id, username, COALESCE(full_name, ''), role_title
		FROM admin_users
		WHERE role = 'navahi_manategh' AND is_active = true
		ORDER BY full_name ASC, id ASC
	`)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []RecipientUserItem
	for rows.Next() {
		var u RecipientUserItem
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.RoleTitle); err == nil {
			list = append(list, u)
		}
	}
	if list == nil {
		list = []RecipientUserItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type ForwardReportRequest struct {
	ReportID           int64    `json:"report_id"`
	RecipientIDs       []int64  `json:"recipient_ids"`
	RecipientUsernames []string `json:"recipient_usernames"`
}

func ForwardReportHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)
	if admin.Role != "superuser" && admin.Role != "edare_kol" && !admin.IsSuperuser {
		http.Error(w, "دسترسی غیرمجاز: تنها اداره کل و مدیر ارشد مجاز به ارجاع خبر هستند.", http.StatusForbidden)
		return
	}

	var req ForwardReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "درخواست نامعتبر است.", http.StatusBadRequest)
		return
	}

	if req.ReportID <= 0 || (len(req.RecipientIDs) == 0 && len(req.RecipientUsernames) == 0) {
		http.Error(w, "لطفاً حداقل یک کارشناس حراست جهت ارجاع خبر انتخاب نمایید.", http.StatusBadRequest)
		return
	}

	var trackingCode string
	err := DB.QueryRow("SELECT tracking_code FROM reports WHERE id = $1", req.ReportID).Scan(&trackingCode)
	if err != nil || trackingCode == "" {
		http.Error(w, "پرونده گزارش مورد نظر یافت نشد.", http.StatusNotFound)
		return
	}

	type validRecip struct {
		id       int64
		username string
		fullName string
	}
	var validRecips []validRecip

	if len(req.RecipientIDs) > 0 {
		recipRows, err := DB.Query(`
			SELECT id, username, full_name
			FROM admin_users
			WHERE id = ANY($1) AND role = 'navahi_manategh' AND is_active = true
		`, pq.Array(req.RecipientIDs))
		if err != nil {
			http.Error(w, "خطای دیتابیس در استعلام کارشناسان", http.StatusInternalServerError)
			return
		}
		defer recipRows.Close()

		for recipRows.Next() {
			var vr validRecip
			if err := recipRows.Scan(&vr.id, &vr.username, &vr.fullName); err == nil {
				validRecips = append(validRecips, vr)
			}
		}
	} else if len(req.RecipientUsernames) > 0 {
		recipRows, err := DB.Query(`
			SELECT id, username, full_name
			FROM admin_users
			WHERE username = ANY($1) AND role = 'navahi_manategh' AND is_active = true
		`, pq.Array(req.RecipientUsernames))
		if err != nil {
			http.Error(w, "خطای دیتابیس در استعلام کارشناسان", http.StatusInternalServerError)
			return
		}
		defer recipRows.Close()

		for recipRows.Next() {
			var vr validRecip
			if err := recipRows.Scan(&vr.id, &vr.username, &vr.fullName); err == nil {
				validRecips = append(validRecips, vr)
			}
		}
	}

	if len(validRecips) == 0 {
		http.Error(w, "هیچ کارشناس معتبری با نقش نواحی و مناطق جهت ارجاع یافت نشد.", http.StatusBadRequest)
		return
	}

	now := time.Now()
	senderDisplayName := admin.FullName
	if strings.TrimSpace(senderDisplayName) == "" {
		senderDisplayName = admin.Username
	}
	if strings.TrimSpace(admin.Title) != "" {
		senderDisplayName = fmt.Sprintf("%s (%s)", senderDisplayName, admin.Title)
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "خطای دیتابیس در شروع تراکنش", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	for _, vr := range validRecips {
		res, err := tx.Exec(`
			INSERT INTO report_forwardings (report_id, forwarded_by, recipient_username, forwarded_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (report_id, recipient_username) DO NOTHING
		`, req.ReportID, admin.Username, vr.username, now)
		if err != nil {
			log.Printf("Error inserting forwarding for report %d to %s: %v", req.ReportID, vr.username, err)
			continue
		}

		rowsAffected, _ := res.RowsAffected()
		if rowsAffected > 0 {

			_, notifErr := tx.Exec(`
				INSERT INTO admin_notifications (notif_type, title, message, target_report_id, target_tracking_code, recipient_username)
				VALUES ($1, $2, $3, $4, $5, $6)
			`,
				"report_forwarded",
				"ارجاع خبر جدید از اداره کل",
				fmt.Sprintf("پرونده محرمانه با شماره پیگیری %s توسط %s به کارتابل شما ارجاع داده شد.", trackingCode, senderDisplayName),
				req.ReportID,
				trackingCode,
				vr.username,
			)
			if notifErr != nil {
				log.Printf("Error inserting forward notification for %s: %v", vr.username, notifErr)
			}
		}
	}

	if err := logAuditTx(tx, admin.Username, admin.Title, fmt.Sprintf("FORWARD_REPORT:%d_TO_%d_RECIPIENTS", req.ReportID, len(validRecips)), &req.ReportID, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "خطای دیتابیس در تایید ارجاع", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":          "success",
		"message":         "گزارش با موفقیت به کارشناسان منتخب ارجاع داده شد.",
		"recipient_count": len(validRecips),
	})
}

type ForwardHistoryItem struct {
	ID                int64  `json:"id"`
	ReportID          int64  `json:"report_id"`
	ForwardedBy       string `json:"forwarded_by"`
	ForwarderName     string `json:"forwarder_name"`
	ForwardedByName   string `json:"forwarded_by_name"`
	ForwarderTitle    string `json:"forwarder_title"`
	ForwardedByTitle  string `json:"forwarded_by_title"`
	RecipientUsername string `json:"recipient_username"`
	RecipientName     string `json:"recipient_name"`
	RecipientTitle    string `json:"recipient_title"`
	ForwardedAt       string `json:"forwarded_at"`
	IsViewed          bool   `json:"is_viewed"`
	ViewedAt          string `json:"viewed_at"`
	ResponseText      string `json:"response_text,omitempty"`
	RespondedAt       string `json:"responded_at,omitempty"`
}

func GetForwardingHistoryHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	reportIDStr := r.URL.Query().Get("id")
	reportID, err := strconv.ParseInt(reportIDStr, 10, 64)
	if err != nil || reportID <= 0 {
		http.Error(w, "شناسه پرونده نامعتبر است.", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, reportID) {
		http.Error(w, "دسترسی غیرمجاز به پرونده گزارش.", http.StatusForbidden)
		return
	}

	rows, err := DB.Query(`
		SELECT rf.id, rf.report_id, rf.forwarded_by,
		       COALESCE(NULLIF(u_f.full_name, ''), rf.forwarded_by) AS forwarder_name,
		       COALESCE(u_f.role_title, '') AS forwarder_title,
		       rf.recipient_username,
		       COALESCE(NULLIF(u_r.full_name, ''), rf.recipient_username) AS recipient_name,
		       COALESCE(u_r.role_title, '') AS recipient_title,
		       rf.forwarded_at, rf.is_viewed, rf.viewed_at,
		       rf.response_text, rf.responded_at
		FROM report_forwardings rf
		LEFT JOIN admin_users u_f ON rf.forwarded_by = u_f.username
		LEFT JOIN admin_users u_r ON rf.recipient_username = u_r.username
		WHERE rf.report_id = $1
		ORDER BY rf.forwarded_at ASC, rf.id ASC
	`, reportID)
	if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var list []ForwardHistoryItem
	for rows.Next() {
		var item ForwardHistoryItem
		var fwdAt time.Time
		var viewedAt sql.NullTime
		var encResp sql.NullString
		var respAt sql.NullTime
		if err := rows.Scan(&item.ID, &item.ReportID, &item.ForwardedBy, &item.ForwarderName, &item.ForwarderTitle,
			&item.RecipientUsername, &item.RecipientName, &item.RecipientTitle,
			&fwdAt, &item.IsViewed, &viewedAt,
			&encResp, &respAt); err == nil {
			if strings.TrimSpace(item.ForwarderName) == "" || item.ForwarderName == item.ForwardedBy {
				if strings.TrimSpace(item.ForwarderTitle) != "" {
					item.ForwarderName = item.ForwarderTitle
				} else {
					item.ForwarderName = "کارشناس اداره کل"
				}
			}
			item.ForwardedByName = item.ForwarderName
			item.ForwardedByTitle = item.ForwarderTitle
			item.ForwardedAt = fwdAt.Format(time.RFC3339)
			if viewedAt.Valid {
				item.ViewedAt = viewedAt.Time.Format(time.RFC3339)
			}
			if encResp.Valid && encResp.String != "" {
				dec, err := DecryptAES(encResp.String)
				if err == nil {
					item.ResponseText = dec
				} else {
					item.ResponseText = "[خطا در رمزگشایی گزارش]"
				}
			}
			if respAt.Valid {
				item.RespondedAt = respAt.Time.Format(time.RFC3339)
			}
			list = append(list, item)
		}
	}
	if list == nil {
		list = []ForwardHistoryItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type AdminNotificationItem struct {
	ID                 int64   `json:"id"`
	NotifType          string  `json:"notif_type"`
	Title              string  `json:"title"`
	Message            string  `json:"message"`
	TargetReportID     *int64  `json:"target_report_id"`
	TargetTrackingCode *string `json:"target_tracking_code"`
	IsRead             bool    `json:"is_read"`
	CreatedAt          string  `json:"created_at"`
}

type AdminNotificationsResponse struct {
	Notifications []AdminNotificationItem `json:"notifications"`
	UnreadCount   int                     `json:"unread_count"`
}

func GetNotificationsHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var resp AdminNotificationsResponse
	resp.Notifications = []AdminNotificationItem{}

	var unreadQuery string
	var rowsQuery string
	var args []interface{}

	if admin.Role == "navahi_manategh" {

		unreadQuery = `
			SELECT COUNT(*) FROM admin_notifications
			WHERE is_read = false
			  AND (
				recipient_username = $1
				OR (recipient_username IS NULL AND notif_type = 'new_reply' AND target_report_id IN (
					SELECT report_id FROM report_forwardings WHERE recipient_username = $1
				))
			  )
		`
		rowsQuery = `
			SELECT id, notif_type, title, message, target_report_id, target_tracking_code, is_read, created_at
			FROM admin_notifications
			WHERE recipient_username = $1
			   OR (recipient_username IS NULL AND notif_type = 'new_reply' AND target_report_id IN (
					SELECT report_id FROM report_forwardings WHERE recipient_username = $1
			   ))
			ORDER BY id DESC
			LIMIT 50
		`
		args = append(args, admin.Username)
	} else {

		unreadQuery = `
			SELECT COUNT(*) FROM admin_notifications
			WHERE is_read = false
			  AND (recipient_username IS NULL OR recipient_username = $1)
		`
		rowsQuery = `
			SELECT id, notif_type, title, message, target_report_id, target_tracking_code, is_read, created_at
			FROM admin_notifications
			WHERE recipient_username IS NULL OR recipient_username = $1
			ORDER BY id DESC
			LIMIT 50
		`
		args = append(args, admin.Username)
	}

	_ = DB.QueryRow(unreadQuery, args...).Scan(&resp.UnreadCount)

	rows, err := DB.Query(rowsQuery, args...)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
		return
	}
	defer rows.Close()

	for rows.Next() {
		var item AdminNotificationItem
		var targetReportID sql.NullInt64
		var targetTrackingCode sql.NullString
		var createdAt time.Time

		if err := rows.Scan(&item.ID, &item.NotifType, &item.Title, &item.Message, &targetReportID, &targetTrackingCode, &item.IsRead, &createdAt); err != nil {
			continue
		}

		if targetReportID.Valid {
			val := targetReportID.Int64
			item.TargetReportID = &val
		}
		if targetTrackingCode.Valid && targetTrackingCode.String != "" {
			code := targetTrackingCode.String
			item.TargetTrackingCode = &code
		}
		item.CreatedAt = createdAt.Format(time.RFC3339)

		resp.Notifications = append(resp.Notifications, item)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating notifications: %v", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func MarkNotificationReadHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)
	var req struct {
		ID int64 `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ID <= 0 {
		http.Error(w, "شناسه اعلان نامعتبر است", http.StatusBadRequest)
		return
	}

	if admin.Role == "navahi_manategh" {
		_, err := DB.Exec(`
			UPDATE admin_notifications
			SET is_read = true
			WHERE id = $1
			  AND (
				recipient_username = $2
				OR (recipient_username IS NULL AND notif_type = 'new_reply' AND target_report_id IN (
					SELECT report_id FROM report_forwardings WHERE recipient_username = $2
				))
			  )
		`, req.ID, admin.Username)
		if err != nil {
			http.Error(w, "خطا در به‌روزرسانی اعلان", http.StatusInternalServerError)
			return
		}
	} else {
		_, err := DB.Exec("UPDATE admin_notifications SET is_read = true WHERE id = $1 AND (recipient_username IS NULL OR recipient_username = $2)", req.ID, admin.Username)
		if err != nil {
			http.Error(w, "خطا در به‌روزرسانی اعلان", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}

func MarkAllNotificationsReadHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	if admin.Role == "navahi_manategh" {
		_, err := DB.Exec(`
			UPDATE admin_notifications
			SET is_read = true
			WHERE is_read = false
			  AND (
				recipient_username = $1
				OR (recipient_username IS NULL AND notif_type = 'new_reply' AND target_report_id IN (
					SELECT report_id FROM report_forwardings WHERE recipient_username = $1
				))
			  )
		`, admin.Username)
		if err != nil {
			http.Error(w, "خطا در علامت‌گذاری اعلان‌ها", http.StatusInternalServerError)
			return
		}
	} else {
		_, err := DB.Exec(`
			UPDATE admin_notifications
			SET is_read = true
			WHERE is_read = false
			  AND (recipient_username IS NULL OR recipient_username = $1)
		`, admin.Username)
		if err != nil {
			http.Error(w, "خطا در علامت‌گذاری اعلان‌ها", http.StatusInternalServerError)
			return
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"success"}`))
}

func gregorianToJalali(gy, gm, gd int) (jy, jm, jd int) {
	gDaysInMonth := []int{31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	jDaysInMonth := []int{31, 31, 31, 31, 31, 31, 30, 30, 30, 30, 30, 29}

	gy2 := gy - 1600
	gm2 := gm - 1
	gd2 := gd - 1

	gDayNo := 365*gy2 + (gy2+3)/4 - (gy2+99)/100 + (gy2+399)/400
	for i := 0; i < gm2; i++ {
		gDayNo += gDaysInMonth[i]
	}
	if gm2 > 1 && ((gy%4 == 0 && gy%100 != 0) || (gy%400 == 0)) {
		gDayNo++
	}
	gDayNo += gd2

	jDayNo := gDayNo - 79

	jNp := jDayNo / 12053
	jDayNo = jDayNo % 12053

	jy = 979 + 33*jNp + 4*(jDayNo/1461)
	jDayNo %= 1461

	if jDayNo >= 366 {
		jy += (jDayNo - 1) / 365
		jDayNo = (jDayNo - 1) % 365
	}

	for i := 0; i < 11 && jDayNo >= jDaysInMonth[i]; i++ {
		jDayNo -= jDaysInMonth[i]
		jm = i + 1
	}
	jm++
	jd = jDayNo + 1
	return
}

func formatJalaliTime(t time.Time) string {
	t = t.Local()
	jy, jm, jd := gregorianToJalali(t.Year(), int(t.Month()), t.Day())
	return fmt.Sprintf("%04d/%02d/%02d - %02d:%02d", jy, jm, jd, t.Hour(), t.Minute())
}

func ExportReportDocxHandler(w http.ResponseWriter, r *http.Request) {
	admin, ok := r.Context().Value("admin").(AdminSession)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	idStr := r.URL.Query().Get("id")
	if idStr == "" {
		http.Error(w, "شناسه گزارش الزامی است", http.StatusBadRequest)
		return
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		http.Error(w, "شناسه گزارش نامعتبر است", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, id) {
		http.Error(w, "دسترسی غیرمجاز به این گزارش", http.StatusForbidden)
		return
	}

	var trackingCode, category, reportType, priority, encDesc string
	var encName, encPhone, encResponse, respondedBy, encNotes sql.NullString
	var createdAt time.Time
	var readAt, respondedAt sql.NullTime
	var isRead bool

	err = DB.QueryRow(`
		SELECT r.tracking_code, r.category, r.report_type, r.priority, r.reporter_name, r.reporter_phone,
		       r.description, r.admin_response, COALESCE(NULLIF(u.full_name, ''), r.responded_by, ''),
		       r.responded_at, r.internal_notes, r.is_read, r.read_at, r.created_at
		FROM reports r
		LEFT JOIN admin_users u ON r.responded_by = u.username
		WHERE r.id = $1
	`, id).Scan(
		&trackingCode, &category, &reportType, &priority, &encName, &encPhone,
		&encDesc, &encResponse, &respondedBy, &respondedAt, &encNotes, &isRead, &readAt, &createdAt,
	)
	if err == sql.ErrNoRows {
		http.Error(w, "گزارش مورد نظر یافت نشد", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "خطا در بازیابی اطلاعات پرونده", http.StatusInternalServerError)
		return
	}

	description := ""
	if dec, err := DecryptAES(encDesc); err == nil {
		description = dec
	} else {
		description = "[خطای رمزگشایی - اصالت محتوا تأیید نشد]"
	}
	reporterName := "ثبت نشده"
	if encName.Valid && encName.String != "" {
		if dec, err := DecryptAES(encName.String); err == nil && dec != "" {
			reporterName = dec
		}
	}
	reporterPhone := "ثبت نشده"
	if encPhone.Valid && encPhone.String != "" {
		if dec, err := DecryptAES(encPhone.String); err == nil && dec != "" {
			reporterPhone = dec
		}
	}

	adminResponse := "فاقد پاسخ کارشناس حراست"
	hasAdminResponse := false
	if encResponse.Valid && strings.TrimSpace(encResponse.String) != "" {
		if dec, err := DecryptAES(encResponse.String); err == nil && strings.TrimSpace(dec) != "" {
			adminResponse = strings.TrimSpace(dec)
			hasAdminResponse = true
		}
	}

	internalNotes := "فاقد یادداشت محرمانه داخلی."
	if encNotes.Valid && encNotes.String != "" {
		if dec, err := DecryptAES(encNotes.String); err == nil && dec != "" {
			internalNotes = dec
		}
	}

	var hasEvidence bool
	var evidenceCount int
	_ = DB.QueryRow("SELECT COUNT(*) FROM evidence_files WHERE report_id = $1", id).Scan(&evidenceCount)
	hasEvidence = evidenceCount > 0

	createdJalali := formatJalaliTime(createdAt)
	printJalali := formatJalaliTime(time.Now())

	priorityLabel := "عادی"
	switch priority {
	case "urgent":
		priorityLabel = "فوری"
	case "critical":
		priorityLabel = "آنی و ویژه"
	}

	reporterTypeLabel := "ناشناس"
	reporterInfo := "محفوظ و ناشناس"
	if reportType != "anonymous" {
		reporterTypeLabel = "با هویت"
		reporterInfo = fmt.Sprintf("%s (%s)", reporterName, reporterPhone)
	}

	evidenceStatus := "فاقد فایل پیوست"
	if hasEvidence {
		evidenceStatus = fmt.Sprintf("دارای فایل پیوست (%d مورد)", evidenceCount)
	}

	readStatus := "در انتظار بررسی"
	if isRead && readAt.Valid {
		readStatus = fmt.Sprintf("بررسی شده در %s", formatJalaliTime(readAt.Time))
	}

	respBy := "ثبت نشده"
	if hasAdminResponse && respondedBy.Valid && strings.TrimSpace(respondedBy.String) != "" {
		respBy = strings.TrimSpace(respondedBy.String)
	}
	respAtStr := "ثبت نشده"
	if hasAdminResponse && respondedAt.Valid {
		respAtStr = formatJalaliTime(respondedAt.Time)
	}

	payload := map[string]interface{}{
		"tracking_code":   trackingCode,
		"date":            createdJalali,
		"print_date":      printJalali,
		"category":        category,
		"priority":        priorityLabel,
		"reporter_type":   reporterTypeLabel,
		"reporter_info":   reporterInfo,
		"evidence_status": evidenceStatus,
		"read_status":     readStatus,
		"description":     description,
		"admin_response":  adminResponse,
		"responded_by":    respBy,
		"responded_at":    respAtStr,
		"internal_notes":  internalNotes,
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		http.Error(w, "خطا در آماده‌سازی داده‌ها", http.StatusInternalServerError)
		return
	}

	templatePath := filepath.Join(".", "templates", "report_template.docx")
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		http.Error(w, "فایل قالب ورد (report_template.docx) یافت نشد.", http.StatusInternalServerError)
		return
	}
	absTemplatePath, err := filepath.Abs(templatePath)
	if err != nil {
		http.Error(w, "فایل قالب ورد یافت نشد.", http.StatusInternalServerError)
		return
	}

	scriptPath := filepath.Join(".", "scripts", "generate_docx.py")
	if _, err := os.Stat(scriptPath); os.IsNotExist(err) {
		http.Error(w, "اسکریپت ایجاد فایل ورد یافت نشد.", http.StatusInternalServerError)
		return
	}
	absScriptPath, err := filepath.Abs(scriptPath)
	if err != nil {
		http.Error(w, "اسکریپت ایجاد فایل ورد یافت نشد.", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, "python", absScriptPath, absTemplatePath)
	cmd.Dir = os.TempDir()
	cmd.Env = getRestrictedSubprocessEnv()
	cmd.Stdin = bytes.NewReader(jsonBytes)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			http.Error(w, "زمان پردازش ایجاد فایل ورد به پایان رسید (Timeout).", http.StatusGatewayTimeout)
			return
		}
		http.Error(w, fmt.Sprintf("خطا در ایجاد فایل ورد: %s", stderr.String()), http.StatusInternalServerError)
		return
	}

	logAudit(admin.Username, admin.Title, "EXPORTED_REPORT_DOCX", &id, getIP(r))

	filename := fmt.Sprintf("dossier_%s.docx", trackingCode)
	w.Header().Set("Content-Type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	w.Header().Set("Content-Length", strconv.Itoa(stdout.Len()))
	w.Write(stdout.Bytes())
}

func getRestrictedSubprocessEnv() []string {
	env := []string{
		"PYTHONDONTWRITEBYTECODE=1",
		"PYTHONNOUSERSITE=1",
	}
	if runtime.GOOS == "windows" {
		for _, key := range []string{"PATH", "SystemRoot", "COMSPEC", "PATHEXT", "TEMP", "TMP"} {
			if val := os.Getenv(key); val != "" {
				env = append(env, key+"="+val)
			}
		}
	} else {
		env = append(env, "PATH=/usr/bin:/bin:/usr/local/bin")
	}
	return env
}

type SubmitRegionalResponseRequest struct {
	ReportID int64  `json:"report_id"`
	Response string `json:"response"`
}

func SubmitRegionalResponseHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	var req SubmitRegionalResponseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "درخواست نامعتبر است.", http.StatusBadRequest)
		return
	}

	req.Response = strings.TrimSpace(stripNul(req.Response))
	if req.ReportID <= 0 || req.Response == "" {
		http.Error(w, "شناسه پرونده و متن گزارش الزامی است.", http.StatusBadRequest)
		return
	}

	if utf8.RuneCountInString(req.Response) > 8000 {
		http.Error(w, "متن گزارش نمی‌تواند بیش از ۸۰۰۰ کاراکتر باشد.", http.StatusBadRequest)
		return
	}

	var fwdID int64
	err := DB.QueryRow(`
		SELECT id FROM report_forwardings
		WHERE report_id = $1 AND recipient_username = $2
	`, req.ReportID, admin.Username).Scan(&fwdID)
	if err != nil {
		http.Error(w, "این پرونده به کارتابل شما ارجاع نشده است یا ارجاع آن لغو گردیده است.", http.StatusForbidden)
		return
	}

	var trackingCode string
	err = DB.QueryRow("SELECT tracking_code FROM reports WHERE id = $1", req.ReportID).Scan(&trackingCode)
	if err != nil {
		http.Error(w, "پرونده یافت نشد.", http.StatusNotFound)
		return
	}

	encResp, err := EncryptAES(req.Response)
	if err != nil {
		http.Error(w, "خطا در رمزنگاری داده‌ها", http.StatusInternalServerError)
		return
	}

	now := time.Now()
	_, err = DB.Exec(`
		UPDATE report_forwardings
		SET response_text = $1,
		    responded_at = $2,
		    is_viewed = true,
		    viewed_at = COALESCE(viewed_at, $2)
		WHERE id = $3
	`, encResp, now, fwdID)
	if err != nil {
		http.Error(w, "خطا در ذخیره پاسخ بررسی", http.StatusInternalServerError)
		return
	}

	_, _ = DB.Exec("UPDATE reports SET updated_at = NOW() WHERE id = $1", req.ReportID)

	senderDisplayName := admin.FullName
	if strings.TrimSpace(senderDisplayName) == "" {
		senderDisplayName = admin.Username
	}
	if strings.TrimSpace(admin.Title) != "" {
		senderDisplayName = fmt.Sprintf("%s (%s)", senderDisplayName, admin.Title)
	}

	_, _ = DB.Exec(`
		INSERT INTO admin_notifications (notif_type, title, message, target_report_id, target_tracking_code, recipient_username)
		VALUES ($1, $2, $3, $4, $5, NULL)
	`,
		"regional_response",
		"ثبت گزارش بررسی ناحیه/منطقه",
		fmt.Sprintf("کارشناس %s گزارش و نتیجه بررسی پرونده محرمانه %s را ارسال نمود.", senderDisplayName, trackingCode),
		req.ReportID,
		trackingCode,
	)

	GlobalSSEBroker.BroadcastRegionalResponse(req.ReportID)

	logAudit(admin.Username, admin.Title, "SUBMIT_REGIONAL_RESPONSE", &req.ReportID, getIP(r))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "گزارش بررسی شما با موفقیت برای اداره کل ارسال گردید.",
	})
}

type RemoveForwardingRequest struct {
	ForwardingID      int64  `json:"forwarding_id"`
	ReportID          int64  `json:"report_id"`
	RecipientUsername string `json:"recipient_username"`
}

func RemoveForwardingHandler(w http.ResponseWriter, r *http.Request) {
	admin := r.Context().Value("admin").(AdminSession)

	if admin.Role == "navahi_manategh" {
		http.Error(w, "شما مجوز لغو ارجاع پرونده را ندارید.", http.StatusForbidden)
		return
	}

	var req RemoveForwardingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "درخواست نامعتبر است.", http.StatusBadRequest)
		return
	}

	var fwdID int64
	var reportID int64
	var recipUser string

	if req.ForwardingID > 0 {
		err := DB.QueryRow(`
			SELECT id, report_id, recipient_username
			FROM report_forwardings
			WHERE id = $1
		`, req.ForwardingID).Scan(&fwdID, &reportID, &recipUser)
		if err != nil {
			http.Error(w, "ارجاع مورد نظر یافت نشد.", http.StatusNotFound)
			return
		}
	} else if req.ReportID > 0 && strings.TrimSpace(req.RecipientUsername) != "" {
		err := DB.QueryRow(`
			SELECT id, report_id, recipient_username
			FROM report_forwardings
			WHERE report_id = $1 AND recipient_username = $2
		`, req.ReportID, strings.TrimSpace(req.RecipientUsername)).Scan(&fwdID, &reportID, &recipUser)
		if err != nil {
			http.Error(w, "ارجاع مورد نظر یافت نشد.", http.StatusNotFound)
			return
		}
	} else {
		http.Error(w, "شناسه ارجاع یا مشخصات پرونده و گیرنده الزامی است.", http.StatusBadRequest)
		return
	}

	if !canAdminAccessReport(admin, reportID) {
		http.Error(w, "دسترسی غیرمجاز به پرونده", http.StatusForbidden)
		return
	}

	tx, err := DB.BeginTx(r.Context(), nil)
	if err != nil {
		http.Error(w, "خطای دیتابیس", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	_, err = tx.Exec("DELETE FROM report_forwardings WHERE id = $1", fwdID)
	if err != nil {
		http.Error(w, "خطا در حذف ارجاع", http.StatusInternalServerError)
		return
	}

	_, _ = tx.Exec(`
		DELETE FROM admin_notifications
		WHERE target_report_id = $1 AND recipient_username = $2
	`, reportID, recipUser)

	_, _ = tx.Exec(`
		DELETE FROM admin_report_views
		WHERE report_id = $1 AND admin_username = $2
	`, reportID, recipUser)

	if err := logAuditTx(tx, admin.Username, admin.Title, fmt.Sprintf("REVOKE_FORWARD_REP_%d_FROM_%s", reportID, recipUser), &reportID, getIP(r)); err != nil {
		http.Error(w, "سیستم ثبت لاگ با خطا مواجه شد. عملیات متوقف گردید.", http.StatusServiceUnavailable)
		return
	}

	if err := tx.Commit(); err != nil {
		http.Error(w, "خطای دیتابیس در ثبت نهایی", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":  "success",
		"message": "ارجاع پرونده با موفقیت لغو شد و دسترسی کاربر به این پرونده مسدود گردید.",
	})
}
