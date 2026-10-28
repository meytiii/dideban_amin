package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

func ReencryptDatabase(db *sql.DB) (int, error) {
	if db == nil {
		return 0, fmt.Errorf("database connection is nil")
	}

	keyRingMu.RLock()
	currentKID := activeKeyID
	keyRingMu.RUnlock()
	boundPrefix := fmt.Sprintf("v3:%s:b:", currentKID)
	unboundPrefix := fmt.Sprintf("v3:%s:f:", currentKID)

	totalMigrated := 0

	repRows, err := db.Query(`
		SELECT id, description, reporter_name, reporter_phone, admin_response, internal_notes, tracking_code
		FROM reports
		ORDER BY id ASC
	`)
	if err != nil {
		return totalMigrated, fmt.Errorf("failed to query reports: %w", err)
	}
	defer repRows.Close()

	type repItem struct {
		id           int64
		desc         string
		name         sql.NullString
		phone        sql.NullString
		response     sql.NullString
		notes        sql.NullString
		trackingCode string
	}
	var repItems []repItem
	for repRows.Next() {
		var it repItem
		if err := repRows.Scan(&it.id, &it.desc, &it.name, &it.phone, &it.response, &it.notes, &it.trackingCode); err == nil {
			repItems = append(repItems, it)
		}
	}
	_ = repRows.Close()

	for _, it := range repItems {
		needsUpdate := false

		newDesc := it.desc
		if !strings.HasPrefix(it.desc, boundPrefix) {
			dec, err := DecryptBoundField("reports", "description", it.id, it.desc)
			if err == nil {
				if enc, err := EncryptBoundField("reports", "description", it.id, dec); err == nil {
					newDesc = enc
					needsUpdate = true
				}
			}
		}

		newName := it.name
		if it.name.Valid && it.name.String != "" && !strings.HasPrefix(it.name.String, boundPrefix) {
			dec, err := DecryptBoundField("reports", "reporter_name", it.id, it.name.String)
			if err == nil {
				if enc, err := EncryptBoundField("reports", "reporter_name", it.id, dec); err == nil {
					newName = sql.NullString{String: enc, Valid: true}
					needsUpdate = true
				}
			}
		}

		newPhone := it.phone
		if it.phone.Valid && it.phone.String != "" && !strings.HasPrefix(it.phone.String, boundPrefix) {
			dec, err := DecryptBoundField("reports", "reporter_phone", it.id, it.phone.String)
			if err == nil {
				if enc, err := EncryptBoundField("reports", "reporter_phone", it.id, dec); err == nil {
					newPhone = sql.NullString{String: enc, Valid: true}
					needsUpdate = true
				}
			}
		}

		newResponse := it.response
		if it.response.Valid && it.response.String != "" && !strings.HasPrefix(it.response.String, boundPrefix) {
			dec, err := DecryptBoundField("reports", "admin_response", it.id, it.response.String)
			if err == nil {
				if enc, err := EncryptBoundField("reports", "admin_response", it.id, dec); err == nil {
					newResponse = sql.NullString{String: enc, Valid: true}
					needsUpdate = true
				}
			}
		}

		newNotes := it.notes
		if it.notes.Valid && it.notes.String != "" && !strings.HasPrefix(it.notes.String, boundPrefix) {
			dec, err := DecryptBoundField("reports", "internal_notes", it.id, it.notes.String)
			if err == nil {
				if enc, err := EncryptBoundField("reports", "internal_notes", it.id, dec); err == nil {
					newNotes = sql.NullString{String: enc, Valid: true}
					needsUpdate = true
				}
			}
		}

		newTrackingCode := it.trackingCode
		plainTracking := it.trackingCode
		if strings.HasPrefix(plainTracking, "v2:") || strings.HasPrefix(plainTracking, "v3:") {
			if dec, err := DecryptAES(plainTracking); err == nil && dec != "" {
				plainTracking = dec
			}
		}
		newTrackingHash := HashTrackingCode(plainTracking)
		if !strings.HasPrefix(it.trackingCode, unboundPrefix) {
			if enc, err := EncryptAES(plainTracking); err == nil {
				newTrackingCode = enc
				needsUpdate = true
			}
		}

		if needsUpdate {
			_, err = db.Exec(`
				UPDATE reports
				SET description = $1, reporter_name = $2, reporter_phone = $3, admin_response = $4, internal_notes = $5,
				    tracking_code = $6, tracking_code_hash = $7
				WHERE id = $8
			`, newDesc, newName, newPhone, newResponse, newNotes, newTrackingCode, newTrackingHash, it.id)
			if err == nil {
				totalMigrated++
			}
		}
	}

	msgRows, err := db.Query(`SELECT id, message_text FROM report_messages ORDER BY id ASC`)
	if err == nil {
		type msgItem struct {
			id  int64
			msg string
		}
		var msgItems []msgItem
		for msgRows.Next() {
			var m msgItem
			if err := msgRows.Scan(&m.id, &m.msg); err == nil {
				msgItems = append(msgItems, m)
			}
		}
		_ = msgRows.Close()

		for _, m := range msgItems {
			if !strings.HasPrefix(m.msg, boundPrefix) {
				dec, err := DecryptBoundField("report_messages", "message", m.id, m.msg)
				if err == nil {
					if enc, err := EncryptBoundField("report_messages", "message", m.id, dec); err == nil {
						if _, err := db.Exec(`UPDATE report_messages SET message_text = $1 WHERE id = $2`, enc, m.id); err == nil {
							totalMigrated++
						}
					}
				}
			}
		}
	}

	fileRows, err := db.Query(`SELECT id, file_path, file_name FROM evidence_files ORDER BY id ASC`)
	if err == nil {
		type fileItem struct {
			id       int64
			encPath  string
			fileName sql.NullString
		}
		var fileItems []fileItem
		for fileRows.Next() {
			var fi fileItem
			if err := fileRows.Scan(&fi.id, &fi.encPath, &fi.fileName); err == nil {
				fileItems = append(fileItems, fi)
			}
		}
		_ = fileRows.Close()

		for _, fi := range fileItems {
			needsUpdate := false
			newPath := fi.encPath
			realPath, err := DecryptBoundField("evidence_files", "file_path", fi.id, fi.encPath)
			if err == nil && !strings.HasPrefix(fi.encPath, boundPrefix) {
				if enc, err := EncryptBoundField("evidence_files", "file_path", fi.id, realPath); err == nil {
					newPath = enc
					needsUpdate = true
				}
			}

			if realPath != "" {
				if fileData, err := os.ReadFile(realPath); err == nil {
					if !bytes.HasPrefix(fileData, []byte("HENC\x03")) {
						if decBlob, err := DecryptFileBytes(fileData); err == nil {
							if encBlob, err := EncryptFileBytes(decBlob); err == nil {
								tmpPath := realPath + ".migtmp"
								if err := os.WriteFile(tmpPath, encBlob, 0600); err == nil {
									if err := os.Rename(tmpPath, realPath); err != nil {
										_ = os.Remove(tmpPath)
									}
								}
							}
						}
					}
				}
			}

			newName := fi.fileName
			if fi.fileName.Valid && fi.fileName.String != "" && !strings.HasPrefix(fi.fileName.String, boundPrefix) {
				decName, err := DecryptBoundField("evidence_files", "file_name", fi.id, fi.fileName.String)
				if err == nil {
					if encName, err := EncryptBoundField("evidence_files", "file_name", fi.id, decName); err == nil {
						newName = sql.NullString{String: encName, Valid: true}
						needsUpdate = true
					}
				}
			}

			if needsUpdate {
				if _, err := db.Exec(`UPDATE evidence_files SET file_path = $1, file_name = $2 WHERE id = $3`, newPath, newName, fi.id); err == nil {
					totalMigrated++
				}
			}
		}
	}

	userRows, err := db.Query(`SELECT id, totp_secret FROM admin_users WHERE totp_secret IS NOT NULL AND totp_secret != ''`)
	if err == nil {
		type userItem struct {
			id     int64
			secret string
		}
		var userItems []userItem
		for userRows.Next() {
			var ui userItem
			if err := userRows.Scan(&ui.id, &ui.secret); err == nil {
				userItems = append(userItems, ui)
			}
		}
		_ = userRows.Close()

		for _, ui := range userItems {
			if !strings.HasPrefix(ui.secret, boundPrefix) {
				dec, err := DecryptBoundField("admin_users", "totp_secret", ui.id, ui.secret)
				if err == nil {
					if enc, err := EncryptBoundField("admin_users", "totp_secret", ui.id, dec); err == nil {
						if _, err := db.Exec(`UPDATE admin_users SET totp_secret = $1 WHERE id = $2`, enc, ui.id); err == nil {
							totalMigrated++
						}
					}
				}
			}
		}
	}

	fwdRows, err := db.Query(`SELECT id, response_text FROM report_forwardings WHERE response_text IS NOT NULL AND response_text != ''`)
	if err == nil {
		type fwdItem struct {
			id       int64
			response string
		}
		var fwdItems []fwdItem
		for fwdRows.Next() {
			var fi fwdItem
			if err := fwdRows.Scan(&fi.id, &fi.response); err == nil {
				fwdItems = append(fwdItems, fi)
			}
		}
		_ = fwdRows.Close()

		for _, fi := range fwdItems {
			if !strings.HasPrefix(fi.response, boundPrefix) {
				dec, err := DecryptBoundField("report_forwardings", "response_text", fi.id, fi.response)
				if err == nil {
					if enc, err := EncryptBoundField("report_forwardings", "response_text", fi.id, dec); err == nil {
						if _, err := db.Exec(`UPDATE report_forwardings SET response_text = $1 WHERE id = $2`, enc, fi.id); err == nil {
							totalMigrated++
						}
					}
				}
			}
		}
	}

	return totalMigrated, nil
}

func reencryptLegacyData(db *sql.DB) error {
	count, err := ReencryptDatabase(db)
	if err != nil {
		return err
	}
	log.Printf("[INFO] Cryptographic database re-encryption completed: %d records updated to bound format.", count)
	return nil
}
