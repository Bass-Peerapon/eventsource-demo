package postgres

import "errors"

var ErrAggregateOutdated = errors.New("aggregate is outdated")
