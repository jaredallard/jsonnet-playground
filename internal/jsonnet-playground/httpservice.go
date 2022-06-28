package jsonnetplayground

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"entgo.io/ent/dialect"
	"github.com/google/go-jsonnet"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
	"github.com/rgst-io/jsonnet-playground/ent"
	"github.com/rgst-io/jsonnet-playground/ent/code"
	"github.com/rgst-io/jsonnet-playground/pkg/web"
	"github.com/sirupsen/logrus"

	// Place any extra imports for your service code here
	///Block(imports)
	// Used by ent
	entsql "entgo.io/ent/dialect/sql"
	logrusmiddleware "github.com/dictyBase/go-middlewares/middlewares/logrus"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/rs/cors"
	///EndBlock(imports)
)

func createDBConnection(databaseUrl string) (*ent.Client, error) {
	db, err := sql.Open("pgx", databaseUrl)
	if err != nil {
		return nil, err
	}

	// Create an ent.Driver from `db`.
	drv := entsql.OpenDB(dialect.Postgres, db)
	return ent.NewClient(ent.Driver(drv)), nil
}

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// NewPublicHTTPService creates a new public http service
func NewPublicHTTPService(conf *Config) *PublicHTTPService {
	return &PublicHTTPService{conf: conf}
}

// PublicHTTPService handles public http service calls
type PublicHTTPService struct {
	db   *ent.Client
	conf *Config
}

func (s *PublicHTTPService) Run(ctx context.Context, log logrus.FieldLogger) error {
	log.Info("creating postgres connection")
	db, err := createDBConnection(s.conf.DatabaseURL)
	if err != nil {
		log.Fatalf("failed opening connection to db: %v", err)
	}
	defer db.Close()
	s.db = db

	if err := db.Schema.Create(ctx); err != nil {
		return errors.Wrap(err, "failed to run migrations")
	}

	r := mux.NewRouter()
	api := r.PathPrefix("/api/v1/").Subrouter()

	// Api middleware
	api.Use(cors.New(cors.Options{
		AllowedOrigins:     []string{"*"},
		AllowedMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowedHeaders:     []string{"Content-Type"},
		OptionsPassthrough: false,
	}).Handler, logrusmiddleware.NewJSONLogger().Middleware)

	// Api Methods
	api.HandleFunc("/code", s.saveCode).Methods("OPTIONS", "POST")
	api.HandleFunc("/code/{id}", s.getCode).Methods("OPTIONS", "GET")
	api.HandleFunc("/execute", s.executeCode).Methods("OPTIONS", "POST")

	// Serve the SPA
	api.PathPrefix("/").Handler(http.NotFoundHandler())

	spa := spaHandler{staticPath: "web/out", indexPath: "index.html"}

	// Serve the Vue app
	r.PathPrefix("/").Handler(spa)

	addr := fmt.Sprintf("%s:%d", s.conf.HTTPAddress, s.conf.HTTPPort)
	serv := &http.Server{
		Handler:      r,
		Addr:         addr,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)

	go func() {
		log.WithField("address", addr).Info("HTTP server started")
		if err := serv.ListenAndServe(); err != nil {
			log.WithError(err).Error("failed to serve")
			// shutdown the sub-context if we fail to start
			// or close for some reason
			cancel()
		}
	}()

	<-ctx.Done()

	log.Info("shutting down HTTP server")

	// create a context with a 20 second timeout to allow the server to gracefully shut down
	ctx, cancel = context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	return serv.Shutdown(ctx)
}

func (s *PublicHTTPService) saveCode(w http.ResponseWriter, rawReq *http.Request) {
	req := web.SaveCodeRequest{}
	err := json.NewDecoder(rawReq.Body).Decode(&req)
	if err != nil {
		web.SendErrorResponse(w, errors.Wrap(err, "failed to decode body"), http.StatusBadRequest)
		return
	}

	if req.Contents == "" {
		web.SendErrorResponse(w, fmt.Errorf("missing contents"), http.StatusBadRequest)
		return
	}

	newCode, err := s.db.Code.Query().Where(code.ContentsEQ(req.Contents)).Only(rawReq.Context())
	if ent.IsNotFound(err) {
		// if it doesn't already exist, create a new entry
		if newCode, err = s.db.Code.Create().
			SetContents(req.Contents).Save(rawReq.Context()); err != nil {
			web.SendErrorResponse(w, errors.Wrap(err, "failed to save"), http.StatusBadRequest)
			return
		}
	} else if err != nil {
		web.SendErrorResponse(w, errors.Wrap(err, "failed to query"), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.SaveCodeResponse{
		ID: newCode.ID.String(),
	})
}

func (s *PublicHTTPService) getCode(w http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)
	sid, ok := vars["id"]
	if !ok {
		web.SendErrorResponse(w, fmt.Errorf("missing id"), http.StatusBadRequest)
		return
	}

	id, err := uuid.Parse(sid)
	if err != nil {
		web.SendErrorResponse(w, errors.Wrap(err, "failed to parse id"), http.StatusBadRequest)
	}

	code, err := s.db.Code.Get(req.Context(), id)
	if err != nil {
		web.SendErrorResponse(w, errors.Wrap(err, "failed to get code"), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.GetCodeResponse{
		Contents: code.Contents,
	})
}

func (s *PublicHTTPService) executeCode(w http.ResponseWriter, rawReq *http.Request) {
	req := web.SaveCodeRequest{}
	if err := json.NewDecoder(rawReq.Body).Decode(&req); err != nil {
		web.SendErrorResponse(w, errors.Wrap(err, "failed to decode body"), http.StatusBadRequest)
		return
	}

	if req.Contents == "" {
		web.SendErrorResponse(w, fmt.Errorf("missing contents"), http.StatusBadRequest)
		return
	}

	vm := jsonnet.MakeVM()
	// The first param is a filename for when an error is output. As we are doing this all in memory
	// having a filename returned for an error would likely just be confusing to a user.
	out, err := vm.EvaluateSnippet("", req.Contents)
	if err != nil {
		web.SendErrorResponse(w, errors.Wrap(err, "failed to evaluate code"), http.StatusBadRequest)
		return
	}

	web.SendResponse(w, web.ExecuteResponse{
		Output: out,
	})
}
