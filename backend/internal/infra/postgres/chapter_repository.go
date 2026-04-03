package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/popul/revisemieux/internal/domain/chapter"
)

// Compile-time check that ChapterRepository implements chapter.Repository.
var _ chapter.Repository = (*ChapterRepository)(nil)

// ChapterRepository implements chapter.Repository using PostgreSQL.
type ChapterRepository struct {
	pool *pgxpool.Pool
}

// NewChapterRepository creates a new ChapterRepository.
func NewChapterRepository(pool *pgxpool.Pool) *ChapterRepository {
	return &ChapterRepository{pool: pool}
}

// --- Chapter ---

func (r *ChapterRepository) FindByID(ctx context.Context, id uuid.UUID) (*chapter.Chapter, error) {
	const q = `
		SELECT id, user_id, subject, class_level, name, pack_id,
		       current_revision_id, archived, is_demo, created_at, updated_at
		FROM chapters WHERE id = $1`

	ch := &chapter.Chapter{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&ch.ID, &ch.UserID, &ch.Subject, &ch.ClassLevel, &ch.Name, &ch.PackID,
		&ch.CurrentRevisionID, &ch.Archived, &ch.IsDemo, &ch.CreatedAt, &ch.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("chapter.Repository.FindByID: %w", chapter.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindByID: %w", err)
	}
	return ch, nil
}

func (r *ChapterRepository) FindByUser(ctx context.Context, userID uuid.UUID, includeArchived bool) ([]*chapter.Chapter, error) {
	q := `
		SELECT id, user_id, subject, class_level, name, pack_id,
		       current_revision_id, archived, is_demo, created_at, updated_at
		FROM chapters WHERE user_id = $1`
	if !includeArchived {
		q += ` AND archived = false`
	}
	q += ` ORDER BY updated_at DESC`

	rows, err := r.pool.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindByUser: %w", err)
	}
	defer rows.Close()

	var result []*chapter.Chapter
	for rows.Next() {
		ch := &chapter.Chapter{}
		if err := rows.Scan(
			&ch.ID, &ch.UserID, &ch.Subject, &ch.ClassLevel, &ch.Name, &ch.PackID,
			&ch.CurrentRevisionID, &ch.Archived, &ch.IsDemo, &ch.CreatedAt, &ch.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("chapter.Repository.FindByUser: scan: %w", err)
		}
		result = append(result, ch)
	}
	return result, rows.Err()
}

func (r *ChapterRepository) Save(ctx context.Context, ch *chapter.Chapter) error {
	const q = `
		INSERT INTO chapters (id, user_id, subject, class_level, name, pack_id,
		                      current_revision_id, archived, is_demo, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		ON CONFLICT (id) DO UPDATE SET
			subject = EXCLUDED.subject,
			class_level = EXCLUDED.class_level,
			name = EXCLUDED.name,
			pack_id = EXCLUDED.pack_id,
			current_revision_id = EXCLUDED.current_revision_id,
			archived = EXCLUDED.archived,
			is_demo = EXCLUDED.is_demo,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, q,
		ch.ID, ch.UserID, ch.Subject, ch.ClassLevel, ch.Name, ch.PackID,
		ch.CurrentRevisionID, ch.Archived, ch.IsDemo, ch.CreatedAt, ch.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("chapter.Repository.Save: %w", err)
	}
	return nil
}

// --- Revision ---

func (r *ChapterRepository) FindRevisionByID(ctx context.Context, id uuid.UUID) (*chapter.Revision, error) {
	const q = `
		SELECT id, chapter_id, revision_number, status, created_at
		FROM chapter_revisions WHERE id = $1`

	rev := &chapter.Revision{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&rev.ID, &rev.ChapterID, &rev.RevisionNumber, &rev.Status, &rev.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("chapter.Repository.FindRevisionByID: %w", chapter.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindRevisionByID: %w", err)
	}
	return rev, nil
}

func (r *ChapterRepository) FindCurrentRevision(ctx context.Context, chapterID uuid.UUID) (*chapter.Revision, error) {
	const q = `
		SELECT cr.id, cr.chapter_id, cr.revision_number, cr.status, cr.created_at
		FROM chapter_revisions cr
		JOIN chapters c ON c.current_revision_id = cr.id
		WHERE c.id = $1`

	rev := &chapter.Revision{}
	err := r.pool.QueryRow(ctx, q, chapterID).Scan(
		&rev.ID, &rev.ChapterID, &rev.RevisionNumber, &rev.Status, &rev.CreatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("chapter.Repository.FindCurrentRevision: %w", chapter.ErrNoRevision)
	}
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindCurrentRevision: %w", err)
	}
	return rev, nil
}

func (r *ChapterRepository) CountRevisionsByChapter(ctx context.Context, chapterID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM chapter_revisions WHERE chapter_id = $1`, chapterID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("chapter.Repository.CountRevisionsByChapter: %w", err)
	}
	return count, nil
}

func (r *ChapterRepository) SaveRevision(ctx context.Context, rev *chapter.Revision) error {
	const q = `
		INSERT INTO chapter_revisions (id, chapter_id, revision_number, status, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (id) DO UPDATE SET
			status = EXCLUDED.status`

	_, err := r.pool.Exec(ctx, q,
		rev.ID, rev.ChapterID, rev.RevisionNumber, rev.Status, rev.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("chapter.Repository.SaveRevision: %w", err)
	}
	return nil
}

// --- Page ---

func (r *ChapterRepository) FindPagesByRevision(ctx context.Context, revisionID uuid.UUID) ([]*chapter.Page, error) {
	const q = `
		SELECT id, revision_id, photo_url, page_order, ocr_status, created_at, updated_at
		FROM pages WHERE revision_id = $1
		ORDER BY page_order ASC`

	rows, err := r.pool.Query(ctx, q, revisionID)
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindPagesByRevision: %w", err)
	}
	defer rows.Close()

	var result []*chapter.Page
	for rows.Next() {
		p := &chapter.Page{}
		if err := rows.Scan(
			&p.ID, &p.RevisionID, &p.PhotoURL, &p.PageOrder, &p.OCRStatus,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("chapter.Repository.FindPagesByRevision: scan: %w", err)
		}
		result = append(result, p)
	}
	return result, rows.Err()
}

func (r *ChapterRepository) SavePage(ctx context.Context, p *chapter.Page) error {
	const q = `
		INSERT INTO pages (id, revision_id, photo_url, page_order, ocr_status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7)
		ON CONFLICT (id) DO UPDATE SET
			photo_url = EXCLUDED.photo_url,
			ocr_status = EXCLUDED.ocr_status,
			updated_at = EXCLUDED.updated_at`

	_, err := r.pool.Exec(ctx, q,
		p.ID, p.RevisionID, p.PhotoURL, p.PageOrder, p.OCRStatus,
		p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("chapter.Repository.SavePage: %w", err)
	}
	return nil
}

// --- Item ---

func (r *ChapterRepository) FindItemByID(ctx context.Context, id uuid.UUID) (*chapter.Item, error) {
	const q = `
		SELECT id, chapter_id, notion_id, revision_id, item_type, term,
		       confidence, validation_required, archived,
		       llm_model_version, prompt_template_version,
		       created_at, updated_at
		FROM items WHERE id = $1`

	item := &chapter.Item{}
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&item.ID, &item.ChapterID, &item.NotionID, &item.RevisionID,
		&item.ItemType, &item.Term,
		&item.Confidence, &item.ValidationRequired, &item.Archived,
		&item.LLMModelVersion, &item.PromptTemplateVersion,
		&item.CreatedAt, &item.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("chapter.Repository.FindItemByID: %w", chapter.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindItemByID: %w", err)
	}

	// Load keywords
	keywords, err := r.loadKeywords(ctx, id)
	if err != nil {
		return nil, err
	}
	item.Keywords = keywords

	// Load steps for PROCEDURE items
	if item.ItemType == chapter.ItemProcedure {
		steps, err := r.loadSteps(ctx, id)
		if err != nil {
			return nil, err
		}
		item.Steps = steps
	}

	return item, nil
}

func (r *ChapterRepository) FindItemsByChapter(ctx context.Context, chapterID uuid.UUID, includeArchived bool) ([]*chapter.Item, error) {
	q := `
		SELECT id, chapter_id, notion_id, revision_id, item_type, term,
		       confidence, validation_required, archived,
		       llm_model_version, prompt_template_version,
		       created_at, updated_at
		FROM items WHERE chapter_id = $1`
	if !includeArchived {
		q += ` AND archived = false`
	}
	q += ` ORDER BY created_at ASC`

	rows, err := r.pool.Query(ctx, q, chapterID)
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindItemsByChapter: %w", err)
	}
	defer rows.Close()

	var items []*chapter.Item
	for rows.Next() {
		item := &chapter.Item{}
		if err := rows.Scan(
			&item.ID, &item.ChapterID, &item.NotionID, &item.RevisionID,
			&item.ItemType, &item.Term,
			&item.Confidence, &item.ValidationRequired, &item.Archived,
			&item.LLMModelVersion, &item.PromptTemplateVersion,
			&item.CreatedAt, &item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("chapter.Repository.FindItemsByChapter: scan: %w", err)
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindItemsByChapter: rows: %w", err)
	}

	// Load keywords and steps for each item
	for _, item := range items {
		keywords, err := r.loadKeywords(ctx, item.ID)
		if err != nil {
			return nil, err
		}
		item.Keywords = keywords

		if item.ItemType == chapter.ItemProcedure {
			steps, err := r.loadSteps(ctx, item.ID)
			if err != nil {
				return nil, err
			}
			item.Steps = steps
		}
	}

	return items, nil
}

func (r *ChapterRepository) SaveItem(ctx context.Context, item *chapter.Item) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("chapter.Repository.SaveItem: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := r.saveItemInTx(ctx, tx, item); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *ChapterRepository) SaveItems(ctx context.Context, items []*chapter.Item) error {
	if len(items) == 0 {
		return nil
	}

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("chapter.Repository.SaveItems: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	for _, item := range items {
		if err := r.saveItemInTx(ctx, tx, item); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *ChapterRepository) saveItemInTx(ctx context.Context, tx pgx.Tx, item *chapter.Item) error {
	const itemQ = `
		INSERT INTO items (id, chapter_id, notion_id, revision_id, item_type, term,
		                   confidence, validation_required, archived,
		                   llm_model_version, prompt_template_version,
		                   created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (id) DO UPDATE SET
			notion_id = EXCLUDED.notion_id,
			item_type = EXCLUDED.item_type,
			term = EXCLUDED.term,
			confidence = EXCLUDED.confidence,
			validation_required = EXCLUDED.validation_required,
			archived = EXCLUDED.archived,
			llm_model_version = EXCLUDED.llm_model_version,
			prompt_template_version = EXCLUDED.prompt_template_version,
			updated_at = EXCLUDED.updated_at`

	_, err := tx.Exec(ctx, itemQ,
		item.ID, item.ChapterID, item.NotionID, item.RevisionID,
		item.ItemType, item.Term,
		item.Confidence, item.ValidationRequired, item.Archived,
		item.LLMModelVersion, item.PromptTemplateVersion,
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("chapter.Repository.SaveItem: %w", err)
	}

	// Upsert keywords: delete-then-insert within the same tx
	if _, err := tx.Exec(ctx, `DELETE FROM item_keywords WHERE item_id = $1`, item.ID); err != nil {
		return fmt.Errorf("chapter.Repository.SaveItem: delete keywords: %w", err)
	}
	for i, kw := range item.Keywords {
		_, err := tx.Exec(ctx,
			`INSERT INTO item_keywords (item_id, keyword, position) VALUES ($1, $2, $3)`,
			item.ID, kw, i,
		)
		if err != nil {
			return fmt.Errorf("chapter.Repository.SaveItem: insert keyword: %w", err)
		}
	}

	// Upsert steps for PROCEDURE items
	if item.ItemType == chapter.ItemProcedure {
		if _, err := tx.Exec(ctx, `DELETE FROM item_steps WHERE item_id = $1`, item.ID); err != nil {
			return fmt.Errorf("chapter.Repository.SaveItem: delete steps: %w", err)
		}
		for _, step := range item.Steps {
			_, err := tx.Exec(ctx,
				`INSERT INTO item_steps (id, item_id, step_order, content) VALUES ($1, $2, $3, $4)`,
				step.ID, item.ID, step.StepOrder, step.Content,
			)
			if err != nil {
				return fmt.Errorf("chapter.Repository.SaveItem: insert step: %w", err)
			}
		}
	}

	return nil
}

// --- Notion ---

func (r *ChapterRepository) FindNotionsByChapter(ctx context.Context, chapterID uuid.UUID) ([]*chapter.Notion, error) {
	const q = `
		SELECT id, chapter_id, name, sort_order, created_at
		FROM notions WHERE chapter_id = $1
		ORDER BY sort_order ASC`

	rows, err := r.pool.Query(ctx, q, chapterID)
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository.FindNotionsByChapter: %w", err)
	}
	defer rows.Close()

	var result []*chapter.Notion
	for rows.Next() {
		n := &chapter.Notion{}
		if err := rows.Scan(&n.ID, &n.ChapterID, &n.Name, &n.SortOrder, &n.CreatedAt); err != nil {
			return nil, fmt.Errorf("chapter.Repository.FindNotionsByChapter: scan: %w", err)
		}
		result = append(result, n)
	}
	return result, rows.Err()
}

func (r *ChapterRepository) SaveNotion(ctx context.Context, n *chapter.Notion) error {
	const q = `
		INSERT INTO notions (id, chapter_id, name, sort_order, created_at)
		VALUES ($1,$2,$3,$4,$5)
		ON CONFLICT (id) DO UPDATE SET
			name = EXCLUDED.name,
			sort_order = EXCLUDED.sort_order`

	_, err := r.pool.Exec(ctx, q, n.ID, n.ChapterID, n.Name, n.SortOrder, n.CreatedAt)
	if err != nil {
		return fmt.Errorf("chapter.Repository.SaveNotion: %w", err)
	}
	return nil
}

// --- Helpers ---

func (r *ChapterRepository) loadKeywords(ctx context.Context, itemID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT keyword FROM item_keywords WHERE item_id = $1 ORDER BY position ASC`, itemID)
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository: loadKeywords: %w", err)
	}
	defer rows.Close()

	var keywords []string
	for rows.Next() {
		var kw string
		if err := rows.Scan(&kw); err != nil {
			return nil, fmt.Errorf("chapter.Repository: loadKeywords: scan: %w", err)
		}
		keywords = append(keywords, kw)
	}
	return keywords, rows.Err()
}

func (r *ChapterRepository) loadSteps(ctx context.Context, itemID uuid.UUID) ([]chapter.ItemStep, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id, item_id, step_order, content FROM item_steps WHERE item_id = $1 ORDER BY step_order ASC`,
		itemID)
	if err != nil {
		return nil, fmt.Errorf("chapter.Repository: loadSteps: %w", err)
	}
	defer rows.Close()

	var steps []chapter.ItemStep
	for rows.Next() {
		var s chapter.ItemStep
		if err := rows.Scan(&s.ID, &s.ItemID, &s.StepOrder, &s.Content); err != nil {
			return nil, fmt.Errorf("chapter.Repository: loadSteps: scan: %w", err)
		}
		steps = append(steps, s)
	}
	return steps, rows.Err()
}
