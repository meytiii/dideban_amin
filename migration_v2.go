package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strings"
)

func reencryptLegacyData(db *sql.DB) error {
	log.Println("[MIGRATION-V2] Starting cryptographic key separation data migration...")

	reportsMigrated := 0
	rows, err := db.Query(`
		SELECT id, description, reporter_name, reporter_phone, admin_response, internal_notes
		FROM reports
	`)
	if err != nil {
		return fmt.Errorf("failed to query reports for migration: %w", err)
	}
	defer rows.Close()

	type reportMigrationItem struct {
		id          int64
		desc        string
		name        sql.NullString
		phone       sql.NullString
		response    sql.NullString
		notes       sql.NullString
		needsUpdate bool
	}

	var reportItems []reportMigrationItem
	for rows.Next() {
		var item reportMigrationItem
		if err := rows.Scan(&item.id, &item.desc, &item.name, &item.phone, &item.response, &item.notes); err != nil {
			return fmt.Errorf("failed to scan report row: %w", err)
		}

		if !strings.HasPrefix(item.desc, "v2:") {
			item.needsUpdate = true
		}
		if item.name.Valid && item.name.String != "" && !strings.HasPrefix(item.name.String, "v2:") {
			item.needsUpdate = true
		}
		if item.phone.Valid && item.phone.String != "" && !strings.HasPrefix(item.phone.String, "v2:") {
			item.needsUpdate = true
		}
		if item.response.Valid && item.response.String != "" && !strings.HasPrefix(item.response.String, "v2:") {
			item.needsUpdate = true
		}
		if item.notes.Valid && item.notes.String != "" && !strings.HasPrefix(item.notes.String, "v2:") {
			item.needsUpdate = true
		}

		if item.needsUpdate {
			reportItems = append(reportItems, item)
		}
	}
	rows.Close()

	for _, item := range reportItems {
		tx, err := db.Begin()
		if err != nil {
			return fmt.Errorf("failed to start report tx: %w", err)
		}

		newDesc := item.desc
		if !strings.HasPrefix(item.desc, "v2:") {
			dec, err := DecryptAES(item.desc)
			if err != nil {
				log.Printf("[MIGRATION-V2] Warning: failed to decrypt report %d description: %v (skipping)", item.id, err)
				tx.Rollback()
				continue
			}
			newDesc, err = EncryptAES(dec)
			if err != nil {
				log.Printf("[MIGRATION-V2] Warning: failed to re-encrypt report %d description: %v (skipping)", item.id, err)
				tx.Rollback()
				continue
			}
		}

		var newName sql.NullString
		if item.name.Valid && item.name.String != "" {
			if strings.HasPrefix(item.name.String, "v2:") {
				newName = item.name
			} else {
				dec, err := DecryptAES(item.name.String)
				if err != nil {
					log.Printf("[MIGRATION-V2] Warning: failed to decrypt report %d name: %v", item.id, err)
				} else if enc, err := EncryptAES(dec); err == nil {
					newName = sql.NullString{String: enc, Valid: true}
				}
			}
		}

		var newPhone sql.NullString
		if item.phone.Valid && item.phone.String != "" {
			if strings.HasPrefix(item.phone.String, "v2:") {
				newPhone = item.phone
			} else {
				dec, err := DecryptAES(item.phone.String)
				if err != nil {
					log.Printf("[MIGRATION-V2] Warning: failed to decrypt report %d phone: %v", item.id, err)
				} else if enc, err := EncryptAES(dec); err == nil {
					newPhone = sql.NullString{String: enc, Valid: true}
				}
			}
		}

		var newResponse sql.NullString
		if item.response.Valid && item.response.String != "" {
			if strings.HasPrefix(item.response.String, "v2:") {
				newResponse = item.response
			} else {
				dec, err := DecryptAES(item.response.String)
				if err != nil {
					log.Printf("[MIGRATION-V2] Warning: failed to decrypt report %d response: %v", item.id, err)
				} else if enc, err := EncryptAES(dec); err == nil {
					newResponse = sql.NullString{String: enc, Valid: true}
				}
			}
		}

		var newNotes sql.NullString
		if item.notes.Valid && item.notes.String != "" {
			if strings.HasPrefix(item.notes.String, "v2:") {
				newNotes = item.notes
			} else {
				dec, err := DecryptAES(item.notes.String)
				if err != nil {
					log.Printf("[MIGRATION-V2] Warning: failed to decrypt report %d notes: %v", item.id, err)
				} else if enc, err := EncryptAES(dec); err == nil {
					newNotes = sql.NullString{String: enc, Valid: true}
				}
			}
		}

		_, err = tx.Exec(`
			UPDATE reports
			SET description = $1, reporter_name = $2, reporter_phone = $3, admin_response = $4, internal_notes = $5
			WHERE id = $6
		`, newDesc, newName, newPhone, newResponse, newNotes, item.id)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update report %d: %w", item.id, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit report %d update: %w", item.id, err)
		}
		reportsMigrated++
	}
	log.Printf("[MIGRATION-V2] Migrated %d reports to v2 encryption.", reportsMigrated)

	messagesMigrated := 0
	msgRows, err := db.Query(`SELECT id, message_text FROM report_messages`)
	if err == nil {
		defer msgRows.Close()
		type msgItem struct {
			id      int64
			message string
		}
		var msgItems []msgItem
		for msgRows.Next() {
			var m msgItem
			if err := msgRows.Scan(&m.id, &m.message); err == nil {
				if !strings.HasPrefix(m.message, "v2:") {
					msgItems = append(msgItems, m)
				}
			}
		}
		msgRows.Close()

		for _, m := range msgItems {
			dec, err := DecryptAES(m.message)
			if err != nil {
				return fmt.Errorf("failed to decrypt message %d: %w", m.id, err)
			}
			enc, err := EncryptAES(dec)
			if err != nil {
				return fmt.Errorf("failed to re-encrypt message %d: %w", m.id, err)
			}
			_, err = db.Exec(`UPDATE report_messages SET message_text = $1 WHERE id = $2`, enc, m.id)
			if err != nil {
				return fmt.Errorf("failed to update message %d: %w", m.id, err)
			}
			messagesMigrated++
		}
	}
	log.Printf("[MIGRATION-V2] Migrated %d messages to v2 encryption.", messagesMigrated)

	filesMigrated := 0
	diskBlobsMigrated := 0
	fileRows, err := db.Query(`SELECT id, file_path, file_name FROM evidence_files`)
	if err == nil {
		defer fileRows.Close()
		type fileItem struct {
			id       int64
			encPath  string
			fileName sql.NullString
		}
		var fItems []fileItem
		for fileRows.Next() {
			var fi fileItem
			if err := fileRows.Scan(&fi.id, &fi.encPath, &fi.fileName); err == nil {
				fItems = append(fItems, fi)
			}
		}
		fileRows.Close()

		for _, fi := range fItems {

			realPath, decErr := DecryptFilePath(fi.encPath)
			if decErr == nil && realPath != "" {

				if fileData, readErr := os.ReadFile(realPath); readErr == nil {
					if !bytes.HasPrefix(fileData, []byte("HENC\x02")) {
						decBlob, decBlobErr := DecryptFileBytes(fileData)
						if decBlobErr == nil {
							encBlob, encBlobErr := EncryptFileBytes(decBlob)
							if encBlobErr == nil {
								tmpPath := realPath + ".migtmp"
								if writeErr := os.WriteFile(tmpPath, encBlob, 0600); writeErr == nil {
									if renErr := os.Rename(tmpPath, realPath); renErr == nil {
										diskBlobsMigrated++
									} else {
										_ = os.Remove(tmpPath)
									}
								}
							}
						}
					}
				}
			}

			needsPathUpdate := !strings.HasPrefix(fi.encPath, "v2:p:")
			needsNameUpdate := fi.fileName.Valid && fi.fileName.String != "" && !strings.HasPrefix(fi.fileName.String, "v2:p:")

			if needsPathUpdate || needsNameUpdate {
				newEncPath := fi.encPath
				if needsPathUpdate && decErr == nil {
					newEncPath, _ = EncryptFilePath(realPath)
				}

				newEncName := fi.fileName
				if needsNameUpdate {
					decName, err := DecryptFilePath(fi.fileName.String)
					if err == nil {
						encName, err := EncryptFilePath(decName)
						if err == nil {
							newEncName = sql.NullString{String: encName, Valid: true}
						}
					}
				}

				_, err = db.Exec(`
					UPDATE evidence_files
					SET file_path = $1, file_name = $2
					WHERE id = $3
				`, newEncPath, newEncName, fi.id)
				if err == nil {
					filesMigrated++
				}
			}
		}
	}
	log.Printf("[MIGRATION-V2] Migrated %d evidence file records and %d disk blobs.", filesMigrated, diskBlobsMigrated)

	log.Println("[MIGRATION-V2] Cryptographic migration completed successfully.")
	return nil
}
