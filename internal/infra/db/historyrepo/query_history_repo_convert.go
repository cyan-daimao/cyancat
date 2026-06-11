package historyrepo

import "cyancat/internal/application/queryservice"

// ToQueryHistoryBO DO -> BO
func ToQueryHistoryBO(do *QueryHistoryDO) *queryservice.QueryHistoryBO {
	if do == nil {
		return nil
	}
	return &queryservice.QueryHistoryBO{
		ID:           do.ID,
		ConnID:       do.ConnID,
		SQL:          do.SQL,
		Status:       do.Status,
		ErrorMessage: do.ErrorMessage,
		RowCount:     do.RowCount,
		DurationMs:   do.Duration,
		ExecutedAt:   do.CreatedAt,
	}
}

// ToQueryHistoryBOs 批量
func ToQueryHistoryBOs(list []QueryHistoryDO) []*queryservice.QueryHistoryBO {
	if len(list) == 0 {
		return make([]*queryservice.QueryHistoryBO, 0)
	}
	result := make([]*queryservice.QueryHistoryBO, 0, len(list))
	for i := range list {
		result = append(result, ToQueryHistoryBO(&list[i]))
	}
	return result
}

// ToQueryHistoryDO BO -> DO
func ToQueryHistoryDO(bo *queryservice.QueryHistoryBO) *QueryHistoryDO {
	if bo == nil {
		return nil
	}
	return &QueryHistoryDO{
		ID:           bo.ID,
		ConnID:       bo.ConnID,
		SQL:          bo.SQL,
		Status:       bo.Status,
		ErrorMessage: bo.ErrorMessage,
		RowCount:     bo.RowCount,
		Duration:     bo.DurationMs,
		CreatedAt:    bo.ExecutedAt,
	}
}
