package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/a-h/templ"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/rs/zerolog/log"
	"golang.org/x/crypto/bcrypt"

	"github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/auth"
	db "github.com/JairoRiver/Bitacora-del-tecnonauta/backend/internal/db/generated"
)

// Handler serves the HTML admin interface.
type Handler struct {
	queries         db.Querier
	authSecret      string
	sessionDuration time.Duration
}

// NewHandler creates a Handler for the admin web UI.
func NewHandler(queries db.Querier, authSecret string, sessionDuration time.Duration) *Handler {
	return &Handler{
		queries:         queries,
		authSecret:      authSecret,
		sessionDuration: sessionDuration,
	}
}

// PostRow is the data shown per row in the dashboard table.
type PostRow struct {
	ID         uuid.UUID
	Title      string
	Date       string
	Slug       string
	Categories string // pre-joined: "Go, Linux"
}

// BlockFormData holds the parsed form fields for one content block.
type BlockFormData struct {
	ID            string // UUID string (server-generated)
	Type          string // "p" | "c" | "i" | "t"
	ParagraphText string
	ParagraphConf []string // ParagraphStyleType values
	CodeLanguage  string
	CodeText      string
	CodeStyle     string
	ImageSrc      string
	ImageCaption  string
	TableValue    string   // raw JSON, e.g. [["a","b"],["1","2"]]
	TableHasHeader bool
	TableStyles   []string // CellStyleType[]
	TableColStyle string   // raw JSON for ColumnStyle[]
}

// PostFormData drives both the create and edit post pages.
type PostFormData struct {
	IsNew             bool
	PostID            string
	Title             string
	Subtitle          string
	Date              string // YYYY-MM-DD
	HeroImage         string
	Slug              string
	Categories        string // comma-separated
	Blocks            []BlockFormData
	InitialBlockOrder string // comma-separated IDs for hidden input
	Error             string
}

// MountPublic registers login/logout (no auth required).
func (h *Handler) MountPublic(r chi.Router) {
	r.Get("/login", h.handleLoginPage)
	r.Post("/login", h.handleLoginFormPost)
	r.Post("/logout", h.handleLogout)
}

// MountProtected registers dashboard and post-CRUD routes (auth required).
func (h *Handler) MountProtected(r chi.Router) {
	r.Get("/", h.handleDashboard)
	r.Get("/posts/new", h.handleNewPostPage)
	r.Post("/posts", h.handleCreatePost)
	r.Get("/posts/{id}/edit", h.handleEditPostPage)
	r.Post("/posts/{id}", h.handleUpdatePost)
	r.Post("/posts/{id}/delete", h.handleDeletePost)
	r.Get("/blocks/fragment/{type}", h.handleBlockFragment)
	r.Get("/api-key", h.handleAPIKeyPage)
	r.Post("/api-key/generate", h.handleAPIKeyGenerate)
	r.Post("/api-key/delete", h.handleAPIKeyDelete)
}

// ── Auth handlers ─────────────────────────────────────────────────────────────

func (h *Handler) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	renderHTML(w, r, LoginPage(""))
}

func (h *Handler) handleLoginFormPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		renderHTML(w, r, LoginPage("Formulario inválido"))
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")

	if username == "" || password == "" {
		renderHTML(w, r, LoginPage("Usuario y contraseña son obligatorios"))
		return
	}

	user, err := h.queries.GetUserByUsername(r.Context(), username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			renderHTML(w, r, LoginPage("Credenciales incorrectas"))
			return
		}
		log.Error().Err(err).Msg("web: get user")
		renderHTML(w, r, LoginPage("Error interno"))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		renderHTML(w, r, LoginPage("Credenciales incorrectas"))
		return
	}

	token, err := auth.CreateToken(h.authSecret, user.Username, h.sessionDuration)
	if err != nil {
		log.Error().Err(err).Msg("web: create token")
		renderHTML(w, r, LoginPage("Error interno"))
		return
	}

	auth.SetSessionCookie(w, token, h.sessionDuration)
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) handleLogout(w http.ResponseWriter, r *http.Request) {
	auth.ClearSessionCookie(w)
	http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
}

// ── Dashboard ─────────────────────────────────────────────────────────────────

func (h *Handler) handleDashboard(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())
	posts, err := h.queries.ListPosts(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("web: list posts")
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	rows := make([]PostRow, 0, len(posts))
	for _, p := range posts {
		cats, _ := h.queries.GetCategoriesByPostID(r.Context(), p.ID)
		catNames := make([]string, len(cats))
		for i, c := range cats {
			catNames[i] = c.Name
		}
		rows = append(rows, PostRow{
			ID:         p.ID,
			Title:      p.Title,
			Date:       p.Date.Format("2006-01-02"),
			Slug:       p.Slug,
			Categories: strings.Join(catNames, ", "),
		})
	}
	renderHTML(w, r, Dashboard(username, rows))
}

// ── Post pages ────────────────────────────────────────────────────────────────

func (h *Handler) handleNewPostPage(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())
	renderHTML(w, r, PostForm(username, PostFormData{
		IsNew: true,
		Date:  time.Now().Format("2006-01-02"),
	}))
}

func (h *Handler) handleEditPostPage(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	post, err := h.queries.GetPostByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.NotFound(w, r)
			return
		}
		log.Error().Err(err).Msg("web: get post by id")
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}

	cats, _ := h.queries.GetCategoriesByPostID(r.Context(), id)
	catNames := make([]string, len(cats))
	for i, c := range cats {
		catNames[i] = c.Name
	}

	dbBlocks, _ := h.queries.GetContentBlocksByPostID(r.Context(), id)
	blocks := make([]BlockFormData, 0, len(dbBlocks))
	for _, b := range dbBlocks {
		blocks = append(blocks, dbBlockToFormData(b))
	}

	form := PostFormData{
		IsNew:      false,
		PostID:     post.ID.String(),
		Title:      post.Title,
		Subtitle:   post.Subtitle,
		Date:       post.Date.Format("2006-01-02"),
		HeroImage:  post.HeroImage,
		Slug:       post.Slug,
		Categories: strings.Join(catNames, ", "),
		Blocks:     blocks,
	}
	form.InitialBlockOrder = blockOrderString(blocks)
	renderHTML(w, r, PostForm(username, form))
}

// ── Post CRUD ─────────────────────────────────────────────────────────────────

func (h *Handler) handleCreatePost(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	form := parsePostForm(r, true, "")
	if err := h.savePost(r, form); err != nil {
		form.Error = err.Error()
		renderHTML(w, r, PostForm(username, form))
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) handleUpdatePost(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())
	postIDStr := chi.URLParam(r, "id")
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	form := parsePostForm(r, false, postIDStr)
	if err := h.savePost(r, form); err != nil {
		form.Error = err.Error()
		renderHTML(w, r, PostForm(username, form))
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

func (h *Handler) handleDeletePost(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := h.queries.DeletePost(r.Context(), id); err != nil {
		log.Error().Err(err).Msg("web: delete post")
		http.Error(w, "Error interno", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/admin/", http.StatusSeeOther)
}

// ── Block fragment ────────────────────────────────────────────────────────────

func (h *Handler) handleBlockFragment(w http.ResponseWriter, r *http.Request) {
	blockType := chi.URLParam(r, "type")
	id := uuid.New().String()
	b := BlockFormData{ID: id, Type: blockType}

	switch blockType {
	case "p":
		renderHTML(w, r, BlockParagraph(b))
	case "c":
		renderHTML(w, r, BlockCode(b))
	case "i":
		renderHTML(w, r, BlockImage(b))
	case "t":
		b.TableValue = `[["col1","col2"],["valor1","valor2"]]`
		renderHTML(w, r, BlockTable(b))
	default:
		http.Error(w, "unknown block type", http.StatusBadRequest)
	}
}

// ── Internal helpers ──────────────────────────────────────────────────────────

func renderHTML(w http.ResponseWriter, r *http.Request, c templ.Component) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := c.Render(r.Context(), w); err != nil {
		log.Error().Err(err).Msg("web: render")
	}
}

func parsePostForm(r *http.Request, isNew bool, postIDStr string) PostFormData {
	form := PostFormData{
		IsNew:      isNew,
		PostID:     postIDStr,
		Title:      r.FormValue("title"),
		Subtitle:   r.FormValue("subtitle"),
		Date:       r.FormValue("date"),
		HeroImage:  r.FormValue("hero_image"),
		Slug:       r.FormValue("slug"),
		Categories: r.FormValue("categories"),
	}

	orderStr := r.FormValue("block_order")
	if orderStr == "" {
		return form
	}
	for _, id := range strings.Split(orderStr, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			continue
		}
		form.Blocks = append(form.Blocks, BlockFormData{
			ID:             id,
			Type:           r.FormValue("block_type_" + id),
			ParagraphText:  r.FormValue("p_text_" + id),
			ParagraphConf:  r.Form["p_conf_"+id],
			CodeLanguage:   r.FormValue("c_lang_" + id),
			CodeText:       r.FormValue("c_text_" + id),
			CodeStyle:      r.FormValue("c_style_" + id),
			ImageSrc:       r.FormValue("i_src_" + id),
			ImageCaption:   r.FormValue("i_caption_" + id),
			TableValue:     r.FormValue("t_value_" + id),
			TableHasHeader: r.FormValue("t_has_header_"+id) == "on",
			TableStyles:    r.Form["t_styles_"+id],
			TableColStyle:  r.FormValue("t_col_style_" + id),
		})
	}
	form.InitialBlockOrder = blockOrderString(form.Blocks)
	return form
}

func (h *Handler) savePost(r *http.Request, form PostFormData) error {
	date, err := time.Parse("2006-01-02", form.Date)
	if err != nil {
		return fmt.Errorf("fecha inválida: %q", form.Date)
	}

	var postID uuid.UUID
	if form.IsNew {
		post, err := h.queries.CreatePost(r.Context(), db.CreatePostParams{
			Title:     form.Title,
			Subtitle:  form.Subtitle,
			Date:      date,
			HeroImage: form.HeroImage,
			Slug:      form.Slug,
		})
		if err != nil {
			return fmt.Errorf("error al crear post: %w", err)
		}
		postID = post.ID
	} else {
		id, err := uuid.Parse(form.PostID)
		if err != nil {
			return fmt.Errorf("ID de post inválido")
		}
		_, err = h.queries.UpdatePost(r.Context(), db.UpdatePostParams{
			ID:        id,
			Title:     form.Title,
			Subtitle:  form.Subtitle,
			Date:      date,
			HeroImage: form.HeroImage,
			Slug:      form.Slug,
		})
		if err != nil {
			return fmt.Errorf("error al actualizar post: %w", err)
		}
		postID = id
	}

	// Replace categories
	if err := h.queries.ClearPostCategories(r.Context(), postID); err != nil {
		return fmt.Errorf("error al limpiar categorías: %w", err)
	}
	for _, name := range parseCategoryList(form.Categories) {
		cat, err := h.queries.GetOrCreateCategory(r.Context(), db.GetOrCreateCategoryParams{
			Name: name,
			Slug: slugify(name),
		})
		if err != nil {
			return fmt.Errorf("error al guardar categoría %q: %w", name, err)
		}
		if err := h.queries.AddPostCategory(r.Context(), db.AddPostCategoryParams{
			PostID:     postID,
			CategoryID: cat.ID,
		}); err != nil {
			return fmt.Errorf("error al asociar categoría: %w", err)
		}
	}

	// Replace content blocks
	if err := h.queries.DeleteContentBlocksByPostID(r.Context(), postID); err != nil {
		return fmt.Errorf("error al limpiar bloques: %w", err)
	}
	for i, b := range form.Blocks {
		value, conf, err := encodeBlock(b)
		if err != nil {
			return fmt.Errorf("bloque %d inválido: %w", i+1, err)
		}
		_, err = h.queries.CreateContentBlock(r.Context(), db.CreateContentBlockParams{
			PostID:   postID,
			Position: int32(i + 1),
			Type:     db.BlockType(b.Type),
			Value:    value,
			Conf:     conf,
		})
		if err != nil {
			return fmt.Errorf("error al guardar bloque %d: %w", i+1, err)
		}
	}
	return nil
}

func encodeBlock(b BlockFormData) (value, conf json.RawMessage, err error) {
	switch b.Type {
	case "p":
		value, err = json.Marshal(b.ParagraphText)
		if err != nil {
			return
		}
		styles := b.ParagraphConf
		if styles == nil {
			styles = []string{}
		}
		conf, err = json.Marshal(styles)

	case "c":
		type codeVal struct {
			Language string `json:"language"`
			Text     string `json:"text"`
		}
		value, err = json.Marshal(codeVal{Language: b.CodeLanguage, Text: b.CodeText})
		if err != nil {
			return
		}
		conf, err = json.Marshal(b.CodeStyle)

	case "i":
		type imageVal struct {
			Src     string `json:"src"`
			Caption string `json:"caption"`
		}
		value, err = json.Marshal(imageVal{Src: b.ImageSrc, Caption: b.ImageCaption})
		if err != nil {
			return
		}
		conf = json.RawMessage(`null`)

	case "t":
		raw := b.TableValue
		if raw == "" {
			raw = "[]"
		}
		if !json.Valid([]byte(raw)) {
			err = fmt.Errorf("table value is not valid JSON")
			return
		}
		value = json.RawMessage(raw)

		type columnStyle struct {
			Header interface{} `json:"header"`
			Style  []string    `json:"style"`
		}
		type tableConf struct {
			HasHeader   bool          `json:"has_header"`
			TableStyles []string      `json:"table_styles"`
			ColumnStyle []columnStyle `json:"column_style"`
		}
		var colStyles []columnStyle
		if cs := strings.TrimSpace(b.TableColStyle); cs != "" && cs != "null" && json.Valid([]byte(cs)) {
			_ = json.Unmarshal([]byte(cs), &colStyles)
		}
		if colStyles == nil {
			colStyles = []columnStyle{}
		}
		tableStyles := b.TableStyles
		if tableStyles == nil {
			tableStyles = []string{}
		}
		conf, err = json.Marshal(tableConf{
			HasHeader:   b.TableHasHeader,
			TableStyles: tableStyles,
			ColumnStyle: colStyles,
		})

	default:
		err = fmt.Errorf("unknown block type: %q", b.Type)
	}
	return
}

func dbBlockToFormData(b db.ContentBlock) BlockFormData {
	fd := BlockFormData{ID: b.ID.String(), Type: string(b.Type)}

	switch b.Type {
	case db.BlockTypeP:
		var text string
		_ = json.Unmarshal(b.Value, &text)
		fd.ParagraphText = text
		var styles []string
		_ = json.Unmarshal(b.Conf, &styles)
		fd.ParagraphConf = styles

	case db.BlockTypeC:
		var cv struct {
			Language string `json:"language"`
			Text     string `json:"text"`
		}
		_ = json.Unmarshal(b.Value, &cv)
		fd.CodeLanguage = cv.Language
		fd.CodeText = cv.Text
		var style string
		_ = json.Unmarshal(b.Conf, &style)
		fd.CodeStyle = style

	case db.BlockTypeI:
		var iv struct {
			Src     string `json:"src"`
			Caption string `json:"caption"`
		}
		_ = json.Unmarshal(b.Value, &iv)
		fd.ImageSrc = iv.Src
		fd.ImageCaption = iv.Caption

	case db.BlockTypeT:
		fd.TableValue = string(b.Value)
		var tc struct {
			HasHeader   bool     `json:"has_header"`
			TableStyles []string `json:"table_styles"`
			ColumnStyle []struct {
				Header interface{} `json:"header"`
				Style  []string    `json:"style"`
			} `json:"column_style"`
		}
		_ = json.Unmarshal(b.Conf, &tc)
		fd.TableHasHeader = tc.HasHeader
		fd.TableStyles = tc.TableStyles
		colRaw, _ := json.Marshal(tc.ColumnStyle)
		fd.TableColStyle = string(colRaw)
	}
	return fd
}

func blockOrderString(blocks []BlockFormData) string {
	ids := make([]string, len(blocks))
	for i, b := range blocks {
		ids[i] = b.ID
	}
	return strings.Join(ids, ",")
}

func parseCategoryList(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func slugify(s string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(s), " ", "-"))
}

// ── API Key handlers ──────────────────────────────────────────────────────────

func (h *Handler) handleAPIKeyPage(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())
	data, err := h.apiKeyPageData(r)
	if err != nil {
		renderHTML(w, r, APIKeyPage(username, APIKeyPageData{Error: err.Error()}))
		return
	}
	renderHTML(w, r, APIKeyPage(username, data))
}

func (h *Handler) handleAPIKeyGenerate(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())

	raw, hash, err := auth.GenerateAPIKey()
	if err != nil {
		log.Error().Err(err).Msg("failed to generate api key")
		renderHTML(w, r, APIKeyPage(username, APIKeyPageData{Error: "Error al generar la clave: " + err.Error()}))
		return
	}

	// Remove any existing key before inserting the new one.
	if err := h.queries.DeleteAPIKey(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to delete existing api key")
		renderHTML(w, r, APIKeyPage(username, APIKeyPageData{Error: "Error al eliminar la clave anterior: " + err.Error()}))
		return
	}

	newKey, err := h.queries.CreateAPIKey(r.Context(), hash)
	if err != nil {
		log.Error().Err(err).Msg("failed to store api key")
		renderHTML(w, r, APIKeyPage(username, APIKeyPageData{Error: "Error al guardar la clave: " + err.Error()}))
		return
	}

	renderHTML(w, r, APIKeyPage(username, APIKeyPageData{
		Active:    true,
		CreatedAt: newKey.CreatedAt,
		NewRawKey: raw,
	}))
}

func (h *Handler) handleAPIKeyDelete(w http.ResponseWriter, r *http.Request) {
	username, _ := auth.UsernameFromContext(r.Context())

	if err := h.queries.DeleteAPIKey(r.Context()); err != nil {
		log.Error().Err(err).Msg("failed to delete api key")
		renderHTML(w, r, APIKeyPage(username, APIKeyPageData{Error: "Error al eliminar la clave: " + err.Error()}))
		return
	}

	http.Redirect(w, r, "/admin/api-key", http.StatusSeeOther)
}

// apiKeyPageData reads the current API key state from the DB.
func (h *Handler) apiKeyPageData(r *http.Request) (APIKeyPageData, error) {
	key, err := h.queries.GetAPIKey(r.Context())
	if errors.Is(err, pgx.ErrNoRows) {
		return APIKeyPageData{Active: false}, nil
	}
	if err != nil {
		return APIKeyPageData{}, err
	}
	return APIKeyPageData{Active: true, CreatedAt: key.CreatedAt}, nil
}
