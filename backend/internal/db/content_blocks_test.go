package db_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

// createBlock es un helper que inserta un bloque y falla el test si hay error.
func createBlock(t *testing.T, q db.Querier, params db.CreateContentBlockParams) db.ContentBlock {
	t.Helper()
	block, err := q.CreateContentBlock(context.Background(), params)
	require.NoError(t, err)
	return block
}

// ── Paragraph block (type "p") ────────────────────────────────────────────

func TestCreateContentBlock_Paragraph(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-block-p")

	value := json.RawMessage(`"Este es un párrafo de prueba"`)
	conf := json.RawMessage(`["bold","italic","color-1"]`)

	block := createBlock(t, q, db.CreateContentBlockParams{
		PostID:   post.ID,
		Position: 0,
		Type:     db.BlockTypeP,
		Value:    value,
		Conf:     conf,
	})

	assert.Equal(t, db.BlockTypeP, block.Type)
	assert.Equal(t, int32(0), block.Position)

	// Verifica round-trip JSONB: el valor recuperado debe ser JSON equivalente
	var gotValue string
	require.NoError(t, json.Unmarshal(block.Value, &gotValue))
	assert.Equal(t, "Este es un párrafo de prueba", gotValue)

	var gotConf []string
	require.NoError(t, json.Unmarshal(block.Conf, &gotConf))
	assert.ElementsMatch(t, []string{"bold", "italic", "color-1"}, gotConf)
}

// ── Code block (type "c") ─────────────────────────────────────────────────

func TestCreateContentBlock_Code(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-block-c")

	type codeValue struct {
		Language string `json:"language"`
		Text     string `json:"text"`
	}
	rawValue, _ := json.Marshal(codeValue{Language: "Go", Text: `fmt.Println("hello")`})
	rawConf := json.RawMessage(`"color-2"`)

	block := createBlock(t, q, db.CreateContentBlockParams{
		PostID:   post.ID,
		Position: 0,
		Type:     db.BlockTypeC,
		Value:    rawValue,
		Conf:     rawConf,
	})

	assert.Equal(t, db.BlockTypeC, block.Type)

	var got codeValue
	require.NoError(t, json.Unmarshal(block.Value, &got))
	assert.Equal(t, "Go", got.Language)
	assert.Equal(t, `fmt.Println("hello")`, got.Text)

	var gotConf string
	require.NoError(t, json.Unmarshal(block.Conf, &gotConf))
	assert.Equal(t, "color-2", gotConf)
}

// ── Image block (type "i") ────────────────────────────────────────────────

func TestCreateContentBlock_Image(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-block-i")

	type imageValue struct {
		Src     string `json:"src"`
		Caption string `json:"caption"`
	}
	rawValue, _ := json.Marshal(imageValue{
		Src:     "https://res.cloudinary.com/demo/image/upload/sample.jpg",
		Caption: "Diagrama de arquitectura",
	})
	// ImageBlock no tiene conf → almacenamos JSON null (no SQL NULL)
	rawConf := json.RawMessage(`null`)

	block := createBlock(t, q, db.CreateContentBlockParams{
		PostID:   post.ID,
		Position: 0,
		Type:     db.BlockTypeI,
		Value:    rawValue,
		Conf:     rawConf,
	})

	assert.Equal(t, db.BlockTypeI, block.Type)

	var got imageValue
	require.NoError(t, json.Unmarshal(block.Value, &got))
	assert.Equal(t, "Diagrama de arquitectura", got.Caption)

	// conf debe seguir siendo JSON null
	assert.Equal(t, json.RawMessage(`null`), block.Conf)
}

// ── Table block (type "t") ────────────────────────────────────────────────

func TestCreateContentBlock_Table(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-block-t")

	// TableValueType: (string | number)[][]
	rawValue := json.RawMessage(`[["Lenguaje","Año"],["Go","2009"],["Rust","2010"]]`)

	// TableConf con column_style
	type columnStyle struct {
		Header string   `json:"header"`
		Style  []string `json:"style"`
	}
	type tableConf struct {
		HasHeader   bool          `json:"has_header"`
		TableStyles []string      `json:"table_styles"`
		ColumnStyle []columnStyle `json:"column_style"`
	}
	rawConf, _ := json.Marshal(tableConf{
		HasHeader:   true,
		TableStyles: []string{"bold"},
		ColumnStyle: []columnStyle{
			{Header: "Lenguaje", Style: []string{"bold"}},
			{Header: "Año", Style: []string{"size-xs"}},
		},
	})

	block := createBlock(t, q, db.CreateContentBlockParams{
		PostID:   post.ID,
		Position: 0,
		Type:     db.BlockTypeT,
		Value:    rawValue,
		Conf:     rawConf,
	})

	assert.Equal(t, db.BlockTypeT, block.Type)

	var rows [][]interface{}
	require.NoError(t, json.Unmarshal(block.Value, &rows))
	require.Len(t, rows, 3)
	assert.Equal(t, "Lenguaje", rows[0][0])

	var gotConf tableConf
	require.NoError(t, json.Unmarshal(block.Conf, &gotConf))
	assert.True(t, gotConf.HasHeader)
	assert.Len(t, gotConf.ColumnStyle, 2)
}

// ── Ordering ──────────────────────────────────────────────────────────────

func TestGetContentBlocksByPostID_OrderedByPosition(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-blocks-order")

	positions := []int32{2, 0, 1}
	for _, pos := range positions {
		createBlock(t, q, db.CreateContentBlockParams{
			PostID:   post.ID,
			Position: pos,
			Type:     db.BlockTypeP,
			Value:    json.RawMessage(`"bloque"`),
			Conf:     json.RawMessage(`[]`),
		})
	}

	blocks, err := q.GetContentBlocksByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	require.Len(t, blocks, 3)

	assert.Equal(t, int32(0), blocks[0].Position)
	assert.Equal(t, int32(1), blocks[1].Position)
	assert.Equal(t, int32(2), blocks[2].Position)
}

// ── Update / Delete ───────────────────────────────────────────────────────

func TestUpdateContentBlock(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-update-block")

	original := createBlock(t, q, db.CreateContentBlockParams{
		PostID:   post.ID,
		Position: 0,
		Type:     db.BlockTypeP,
		Value:    json.RawMessage(`"original"`),
		Conf:     json.RawMessage(`[]`),
	})

	updated, err := q.UpdateContentBlock(context.Background(), db.UpdateContentBlockParams{
		ID:       original.ID,
		Position: 1,
		Type:     db.BlockTypeC,
		Value:    json.RawMessage(`{"language":"SQL","text":"SELECT 1"}`),
		Conf:     json.RawMessage(`"color-3"`),
	})
	require.NoError(t, err)

	assert.Equal(t, original.ID, updated.ID)
	assert.Equal(t, db.BlockTypeC, updated.Type)
	assert.Equal(t, int32(1), updated.Position)
}

func TestDeleteContentBlock(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-del-block")

	block := createBlock(t, q, db.CreateContentBlockParams{
		PostID:   post.ID,
		Position: 0,
		Type:     db.BlockTypeP,
		Value:    json.RawMessage(`"borrar"`),
		Conf:     json.RawMessage(`[]`),
	})

	err := q.DeleteContentBlock(context.Background(), block.ID)
	require.NoError(t, err)

	remaining, err := q.GetContentBlocksByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	assert.Empty(t, remaining)
}

func TestDeleteContentBlocksByPostID(t *testing.T) {
	q := newQuerier(t)
	post := createPost(t, q, "post-del-all-blocks")

	for i := int32(0); i < 3; i++ {
		createBlock(t, q, db.CreateContentBlockParams{
			PostID:   post.ID,
			Position: i,
			Type:     db.BlockTypeP,
			Value:    json.RawMessage(`"bloque"`),
			Conf:     json.RawMessage(`[]`),
		})
	}

	err := q.DeleteContentBlocksByPostID(context.Background(), post.ID)
	require.NoError(t, err)

	blocks, err := q.GetContentBlocksByPostID(context.Background(), post.ID)
	require.NoError(t, err)
	assert.Empty(t, blocks, "todos los bloques del post deben haberse eliminado")
}
