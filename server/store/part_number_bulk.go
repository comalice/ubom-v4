package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"strconv"

	ubom "ubom-v4"
)

func scanPartNumberRows(rows *sql.Rows) ([]ubom.PartNumber, map[int64]int, error) {
	parts := []ubom.PartNumber{}
	partsByID := map[int64]int{}
	for rows.Next() {
		var part ubom.PartNumber
		var id int64
		var attributeData []byte
		if err := rows.Scan(&id, &part.Value, &part.SeqDefID, &part.TaxonomyDefID, &part.TaxonomyNodeID, &attributeData, &part.Status); err != nil {
			return nil, nil, err
		}
		part.ID = ubom.PartNumberID(strconv.FormatInt(id, 10))
		if err := json.Unmarshal(attributeData, &part.Attributes); err != nil {
			return nil, nil, err
		}
		parts = append(parts, part)
		partsByID[id] = len(parts) - 1
	}
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}
	return parts, partsByID, nil
}

func scanPartRevisionRows(rows *sql.Rows, parts []ubom.PartNumber, partsByID map[int64]int) (map[ubom.PartRevisionID]string, error) {
	revisionSummaries := map[ubom.PartRevisionID]string{}
	for rows.Next() {
		var partID, id int64
		var revision string
		if err := rows.Scan(&partID, &id, &revision); err != nil {
			return nil, err
		}
		partIndex, ok := partsByID[partID]
		if !ok {
			return nil, errors.New("revision references a part outside the list")
		}
		revisionID := ubom.PartRevisionID(strconv.FormatInt(id, 10))
		parts[partIndex].PartRevisionID = append(parts[partIndex].PartRevisionID, revisionID)
		revisionSummaries[revisionID] = revision
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return revisionSummaries, nil
}
