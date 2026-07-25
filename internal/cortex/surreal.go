package cortex

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// SurrealConfig wires a SurrealStore. The store talks to a SurrealDB server over its
// HTTP /sql endpoint (no SDK - we own the wire, like the LLM provider). Vector recall
// uses an HNSW index over embeddings from the Embedder; lexical recall uses a BM25
// full-text index; Search fuses the two with Reciprocal Rank Fusion (E6 / ADR-002).
type SurrealConfig struct {
	Addr      string   // e.g. http://127.0.0.1:8000
	Namespace string   // default "mimir"
	Database  string   // default "mimir"
	User      string   // default "root"
	Pass      string   // default "root"
	Dimension int      // embedding dimension (must match the Embedder); default 768
	Embedder  Embedder // optional; nil => Search is full-text only
}

// surrealClient is a thin SurrealDB HTTP /sql client.
type surrealClient struct {
	addr string
	auth string // "Basic <base64(user:pass)>"
	http *http.Client
}

type stmtResult struct {
	Result json.RawMessage `json:"result"`
	Status string          `json:"status"`
}

func (c *surrealClient) client() *http.Client {
	if c.http != nil {
		return c.http
	}
	return http.DefaultClient
}

// exec runs one or more SurrealQL statements. ns/db may be empty (root / namespace
// context) for DEFINE NAMESPACE / DEFINE DATABASE.
func (c *surrealClient) exec(ctx context.Context, ns, db, sql string) ([]stmtResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, strings.TrimRight(c.addr, "/")+"/sql", strings.NewReader(sql))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "text/plain")
	req.Header.Set("Authorization", c.auth)
	if ns != "" {
		req.Header.Set("surreal-ns", ns)
	}
	if db != "" {
		req.Header.Set("surreal-db", db)
	}
	resp, err := c.client().Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("surreal: HTTP %d: %s", resp.StatusCode, string(raw))
	}
	var results []stmtResult
	if err := json.Unmarshal(raw, &results); err != nil {
		return nil, fmt.Errorf("surreal: decode response: %w", err)
	}
	return results, nil
}

// SurrealStore is a Cortex store backed by SurrealDB (graph + vector + document).
type SurrealStore struct {
	cli      *surrealClient
	ns, db   string
	dim      int
	embedder Embedder
}

// NewSurrealStore connects to SurrealDB and ensures the schema exists.
func NewSurrealStore(ctx context.Context, cfg SurrealConfig) (*SurrealStore, error) {
	if cfg.Addr == "" {
		cfg.Addr = "http://127.0.0.1:8000"
	}
	if cfg.Namespace == "" {
		cfg.Namespace = "mimir"
	}
	if cfg.Database == "" {
		cfg.Database = "mimir"
	}
	if cfg.User == "" {
		cfg.User = "root"
	}
	if cfg.Pass == "" {
		cfg.Pass = "root"
	}
	if cfg.Dimension == 0 {
		cfg.Dimension = 768
	}
	s := &SurrealStore{
		cli: &surrealClient{
			addr: cfg.Addr,
			auth: "Basic " + basicAuth(cfg.User, cfg.Pass),
		},
		ns:       cfg.Namespace,
		db:       cfg.Database,
		dim:      cfg.Dimension,
		embedder: cfg.Embedder,
	}
	if err := s.init(ctx); err != nil {
		return nil, err
	}
	return s, nil
}

func basicAuth(user, pass string) string {
	return base64Std(user + ":" + pass)
}

func base64Std(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

// init defines the namespace, database, and the neuron/synapse/engram schema with the
// HNSW vector index + BM25 full-text index.
func (s *SurrealStore) init(ctx context.Context) error {
	if _, err := s.cli.exec(ctx, "", "", "DEFINE NAMESPACE IF NOT EXISTS "+s.ns+";"); err != nil {
		return fmt.Errorf("surreal: define namespace: %w", err)
	}
	if _, err := s.cli.exec(ctx, s.ns, "", "DEFINE DATABASE IF NOT EXISTS "+s.db+";"); err != nil {
		return fmt.Errorf("surreal: define database: %w", err)
	}
	schema := fmt.Sprintf(`DEFINE TABLE neuron SCHEMAFULL;
DEFINE FIELD kind ON neuron TYPE string;
DEFINE FIELD layer ON neuron TYPE string DEFAULT "";
DEFINE FIELD title ON neuron TYPE string DEFAULT "";
DEFINE FIELD content ON neuron TYPE string;
DEFINE FIELD embedding ON neuron TYPE option<array<float>>;
DEFINE FIELD decay ON neuron TYPE float DEFAULT 1.0;
DEFINE FIELD created_at ON neuron TYPE datetime DEFAULT time::now();
DEFINE INDEX neuron_hnsw ON neuron FIELDS embedding HNSW DIMENSION %d DIST COSINE;
DEFINE ANALYZER ascii_an TOKENIZERS class, blank FILTERS lowercase, ascii;
DEFINE INDEX neuron_fts ON neuron FIELDS content FULLTEXT ANALYZER ascii_an BM25;
DEFINE TABLE synapse SCHEMAFULL;
DEFINE FIELD in ON synapse TYPE record<neuron>;
DEFINE FIELD out ON synapse TYPE record<neuron>;
DEFINE FIELD kind ON synapse TYPE string;
DEFINE FIELD weight ON synapse TYPE float DEFAULT 1.0;
DEFINE TABLE engram SCHEMAFULL;
DEFINE FIELD neuron ON engram TYPE record<neuron>;
DEFINE FIELD strength ON engram TYPE float DEFAULT 1.0;
DEFINE FIELD consolidated ON engram TYPE bool DEFAULT false;`, s.dim)
	results, err := s.cli.exec(ctx, s.ns, s.db, schema)
	if err != nil {
		return fmt.Errorf("surreal: schema: %w", err)
	}
	for _, r := range results {
		if r.Status == "ERR" {
			return fmt.Errorf("surreal: schema statement failed: %s", strings.Trim(string(r.Result), `"`))
		}
	}
	return nil
}

// neuronRecord is the JSON shape of a neuron row in SurrealDB.
type neuronRecord struct {
	ID        string    `json:"id"`
	Kind      string    `json:"kind"`
	Layer     string    `json:"layer"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	Embedding []float64 `json:"embedding"`
	Decay     float64   `json:"decay"`
}

func (r neuronRecord) toNeuron() Neuron {
	return Neuron{
		ID: r.ID, Kind: NeuronKind(r.Kind), Layer: r.Layer, Title: r.Title,
		Content: r.Content, Embedding: r.Embedding, Decay: r.Decay,
	}
}

// queryNeurons runs a single SELECT and decodes the rows into Neurons.
func (s *SurrealStore) queryNeurons(ctx context.Context, sql string) ([]Neuron, error) {
	results, err := s.cli.exec(ctx, s.ns, s.db, sql)
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return nil, nil
	}
	r := results[len(results)-1]
	if r.Status == "ERR" {
		return nil, fmt.Errorf("surreal: %s", strings.Trim(string(r.Result), `"`))
	}
	var recs []neuronRecord
	if err := json.Unmarshal(r.Result, &recs); err != nil {
		return nil, err
	}
	out := make([]Neuron, 0, len(recs))
	for _, rec := range recs {
		out = append(out, rec.toNeuron())
	}
	return out, nil
}

// PutNeuron stores a neuron (embedding it first if an Embedder is configured and no
// embedding was supplied) and returns its record id.
func (s *SurrealStore) PutNeuron(ctx context.Context, n Neuron) (string, error) {
	emb := n.Embedding
	if s.embedder != nil && len(emb) == 0 && n.Content != "" {
		if v, err := s.embedder.Embed(ctx, n.Content); err == nil && len(v) == s.dim {
			emb = v
		}
	}
	content := neuronContent(n, emb)
	var sql string
	if n.ID != "" {
		sql = "UPSERT " + recordRef(n.ID) + " CONTENT " + content + " RETURN AFTER;"
	} else {
		sql = "CREATE neuron CONTENT " + content + " RETURN AFTER;"
	}
	results, err := s.cli.exec(ctx, s.ns, s.db, sql)
	if err != nil {
		return "", err
	}
	if len(results) == 0 || results[0].Status == "ERR" {
		msg := ""
		if len(results) > 0 {
			msg = strings.Trim(string(results[0].Result), `"`)
		}
		return "", fmt.Errorf("surreal: put neuron: %s", msg)
	}
	var recs []neuronRecord
	if err := json.Unmarshal(results[0].Result, &recs); err != nil {
		return "", err
	}
	if len(recs) == 0 {
		return "", fmt.Errorf("surreal: put neuron: no record returned")
	}
	return recs[0].ID, nil
}

// GetNeuron returns a neuron by its record id.
func (s *SurrealStore) GetNeuron(ctx context.Context, id string) (Neuron, error) {
	ns, err := s.queryNeurons(ctx, "SELECT * FROM "+recordRef(id)+";")
	if err != nil {
		return Neuron{}, err
	}
	if len(ns) == 0 {
		return Neuron{}, nil
	}
	return ns[0], nil
}

// Search recalls neurons by hybrid ranking: vector similarity (HNSW over embeddings)
// fused with BM25 full-text via Reciprocal Rank Fusion. Without an Embedder it falls
// back to full-text only.
func (s *SurrealStore) Search(ctx context.Context, query string, limit int) ([]Neuron, error) {
	if limit <= 0 {
		limit = 5
	}
	var vecHits, txtHits []Neuron
	if s.embedder != nil {
		if vec, err := s.embedder.Embed(ctx, query); err == nil && len(vec) == s.dim {
			vecHits, _ = s.vectorSearch(ctx, vec, limit*2)
		}
	}
	txtHits, _ = s.fulltextSearch(ctx, query, limit*2)
	fused := rrfFuse(vecHits, txtHits)
	if len(fused) > limit {
		fused = fused[:limit]
	}
	return fused, nil
}

func (s *SurrealStore) vectorSearch(ctx context.Context, vec []float64, limit int) ([]Neuron, error) {
	ef := limit * 4
	if ef < 40 {
		ef = 40
	}
	q := fmt.Sprintf("SELECT *, vector::distance::knn() AS dist FROM neuron WHERE embedding <|%d,%d|> %s ORDER BY dist LIMIT %d;",
		limit, ef, vectorLiteral(vec), limit)
	return s.queryNeurons(ctx, q)
}

func (s *SurrealStore) fulltextSearch(ctx context.Context, query string, limit int) ([]Neuron, error) {
	q := fmt.Sprintf("SELECT * FROM neuron WHERE content @0@ %s LIMIT %d;", surqlString(query), limit)
	return s.queryNeurons(ctx, q)
}

// Relate records a synapse (graph edge) between two neurons.
func (s *SurrealStore) Relate(ctx context.Context, syn Synapse) error {
	q := fmt.Sprintf("RELATE %s->synapse->%s SET kind=%s, weight=1.0;", recordRef(syn.From), recordRef(syn.To), surqlString(syn.Kind))
	_, err := s.cli.exec(ctx, s.ns, s.db, q)
	return err
}

// Remember hardens a neuron into a durable engram.
func (s *SurrealStore) Remember(ctx context.Context, e Engram) error {
	q := fmt.Sprintf("CREATE engram SET neuron=%s, strength=%s, consolidated=false;", recordRef(e.NeuronID), strconv.FormatFloat(e.Strength, 'g', -1, 64))
	_, err := s.cli.exec(ctx, s.ns, s.db, q)
	return err
}

// All returns every neuron (for the /memory endpoint + tests).
func (s *SurrealStore) All() []Neuron {
	ns, err := s.queryNeurons(context.Background(), "SELECT * FROM neuron ORDER BY id;")
	if err != nil {
		return nil
	}
	return ns
}

// rrfFuse merges two ranked result lists with Reciprocal Rank Fusion (k=60).
func rrfFuse(a, b []Neuron) []Neuron {
	const k = 60
	scores := map[string]float64{}
	byID := map[string]Neuron{}
	var order []string
	add := func(hits []Neuron) {
		for rank, n := range hits {
			if _, ok := byID[n.ID]; !ok {
				order = append(order, n.ID)
			}
			scores[n.ID] += 1.0 / float64(k+rank+1)
			byID[n.ID] = n
		}
	}
	add(a)
	add(b)
	sort.SliceStable(order, func(i, j int) bool { return scores[order[i]] > scores[order[j]] })
	out := make([]Neuron, 0, len(order))
	for _, id := range order {
		out = append(out, byID[id])
	}
	return out
}

// neuronContent builds a SurrealDB CONTENT object (JSON) for a neuron.
func neuronContent(n Neuron, emb []float64) string {
	m := map[string]any{
		"kind":    string(n.Kind),
		"layer":   n.Layer,
		"title":   n.Title,
		"content": n.Content,
		"decay":   n.Decay,
	}
	if len(emb) > 0 {
		m["embedding"] = emb
	}
	b, _ := json.Marshal(m)
	return string(b)
}

// recordRef turns an id into a safe SurrealDB record reference (neuron:...).
func recordRef(id string) string {
	if strings.Contains(id, ":") {
		return id
	}
	return "neuron:" + id
}

// vectorLiteral formats a float slice as a SurrealDB array literal.
func vectorLiteral(v []float64) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = strconv.FormatFloat(x, 'g', -1, 64)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// surqlString renders a Go string as a safely-escaped SurrealDB string literal.
func surqlString(s string) string {
	b, _ := json.Marshal(s)
	return string(b)
}
