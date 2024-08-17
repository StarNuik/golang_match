package model

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/starnuik/golang_match/pkg/schema"
)

func NewUserQueuePostgres(cfg GridConfig, db *pgxpool.Pool) UserQueue {
	return &pgUserQueue{
		GridConfig: cfg,
		db:         db,
	}
}

type pgUserQueue struct {
	GridConfig
	db *pgxpool.Pool
}

// Add implements UserQueue.
func (m *pgUserQueue) Add(ctx context.Context, user *QueuedUser) error {
	idx := toIndex(user, &m.GridConfig)

	tag, err := m.db.Exec(ctx, `
		insert into UserQueue
			(Name, Skill, Latency, QueuedAt, PosS, PosL)
		values
			($1, $2, $3, $4, $5, $6)`,
		user.Name, user.Skill, user.Latency, user.QueuedAt, idx.S, idx.L)

	if err != nil {
		return err
	}
	if !tag.Insert() {
		log.Printf("tag != INSERT, tag(%s)\n", tag.String())
	}
	if tag.RowsAffected() != 1 {
		log.Printf("tag rows != 1, tag(%s)\n", tag.String())
	}

	return nil
}

func (m *pgUserQueue) GetBin(ctx context.Context, idx BinIdx) ([]*QueuedUser, error) {
	rows, err := m.db.Query(ctx, `
		select Name, Skill, Latency, QueuedAt
		from UserQueue
		where PosS = $1 and PosL = $2`,
		idx.S, idx.L)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bin := []*QueuedUser{}
	for rows.Next() {
		user := QueuedUser{}
		err := rows.Scan(&user.Name, &user.Skill, &user.Latency, &user.QueuedAt)
		if err != nil {
			return nil, err
		}
		bin = append(bin, &user)
	}

	return bin, nil
}

func (m *pgUserQueue) GetBins(ctx context.Context, lo BinIdx, hi BinIdx, minWait time.Duration) ([]*QueuedUser, error) {
	// todo: this might not be a great idea
	// todo: different func calls will have different now-s
	now := time.Now().UTC()
	before := now.Add(-minWait)

	rows, err := m.db.Query(ctx, `
		select Name, Skill, Latency, QueuedAt
		from UserQueue
		where
			PosS >= $1 and PosL >= $2 and
			PosS <= $3 and PosL <= $4 and
			QueuedAt < $5`,
		lo.S, lo.L, hi.S, hi.L, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	bin := []*QueuedUser{}
	for rows.Next() {
		user := QueuedUser{}
		err := rows.Scan(&user.Name, &user.Skill, &user.Latency, &user.QueuedAt)
		if err != nil {
			return nil, err
		}
		bin = append(bin, &user)
	}

	return bin, nil
}

func (m *pgUserQueue) Remove(ctx context.Context, users []string) error {
	// https://github.com/jackc/pgx/issues/108#issuecomment-160804629
	tag, err := m.db.Exec(ctx, `
		delete from UserQueue
		where Name = any ($1)`,
		users)
	if err != nil {
		return err
	}
	if !tag.Delete() {
		log.Printf("tag != DELETE, tag(%s)\n", tag.String())
	}
	if tag.RowsAffected() != int64(len(users)) {
		log.Printf("tag rows != len(toRemove), tag(%s), len(%d)\n", tag.String(), len(users))
	}
	return nil
}

func (m *pgUserQueue) Count(ctx context.Context) (int, error) {
	row := m.db.QueryRow(ctx, `
		select count(*)
		from UserQueue`)

	count := 0
	err := row.Scan(&count)
	if err != nil {
		return -1, err
	}

	return count, nil
}

func (m *pgUserQueue) Parse(req *schema.QueueUserRequest) (*QueuedUser, error) {
	return parse(req)
}
