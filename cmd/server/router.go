package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"migrated-app/internal/calc/bati2018"
	"migrated-app/internal/calc/nonbati2018"
	"migrated-app/internal/dispatch"
	"migrated-app/internal/layout/zes"
	"migrated-app/internal/rates"
)

// batiDecoder adapts the ZES return area (slot 2 = XBXTDDIR, slot 4 =
// XBXTDSR) to the bati2018 rate set.
//
// MIGRATION_NOTE: the TEOM rates (TAU-C-PBBOM*) live in the COM segment
// (slot 3); its accessors are not wired here, so Pbbom stays zero.
type batiDecoder struct{}

func (batiDecoder) Decode(z *rates.ZES) bati2018.RateSet {
	dir := zes.DIR(z[1])
	sr := zes.SR(z[3])
	u := func(v uint64, _ bool) int64 { return int64(v) }
	return bati2018.RateSet{
		SrJan:       sr.JanRaw(),
		SrDepdir:    sr.Depdir(),
		SrCcocom:    sr.Ccocom(),
		SrCcoifp:    sr.Ccoifp(),
		TaucomB:     u(sr.TaucomB()),
		TaudepB:     u(dir.TaudepB()),
		Ptbtas:      u(dir.Ptbtas()),
		TausyndsfpB: u(sr.TausyndsfpB()),
		TaucudfpvnB: u(sr.TaucudfpvnB()),
		TautseB:     u(sr.TautseB()),
		Ptbtgp:      u(sr.Ptbtgp()),
		Ptbgem:      u(sr.Ptbgem()),
	}
}

// nonBatiDecoder adapts the ZES return area to the nonbati2018 rate set.
type nonBatiDecoder struct{}

func (nonBatiDecoder) Decode(z *rates.ZES) nonbati2018.RateSet {
	dir := zes.DIR(z[1])
	sr := zes.SR(z[3])
	u := func(v uint64, _ bool) int64 { return int64(v) }
	return nonbati2018.RateSet{
		SrJan:        sr.JanRaw(),
		SrDepdir:     sr.Depdir(),
		SrCcocom:     sr.Ccocom(),
		SrCcoifp:     sr.Ccoifp(),
		TaucomNB:     u(sr.TaucomNB()),
		TausyndsfpNB: u(sr.TausyndsfpNB()),
		TaucudfpvnNB: u(sr.TaucudfpvnNB()),
		Pnbgem:       u(sr.Pnbgem()),
		TautseNB:     u(sr.TautseNB()),
		Pnbtgp:       u(sr.Pnbtgp()),
		Pnbtxa:       u(sr.Pnbtxa()),
		Pnbcaa:       u(sr.Pnbcaa()),
		TauchagrNB:   u(dir.TauchagrNB()),
		TausmagrNB:   u(dir.TausmagrNB()),
	}
}

type dispatchRequest struct {
	Entree string `json:"entree"`
	CR     int    `json:"cr"`
	RC     int    `json:"rc"`
}

type dispatchResponse struct {
	Sortie string `json:"sortie"`
	CR     int    `json:"cr"`
	RC     int    `json:"rc"`
	Error  string `json:"error,omitempty"`
}

type dispatchFunc func(ctx context.Context, entree, sortie []byte, cr, rc int) (int, int, error)

func dispatchHandler(fn dispatchFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dispatchRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, dispatchResponse{Error: "invalid JSON body: " + err.Error()})
			return
		}
		sortie := make([]byte, dispatch.RecordLen)
		for i := range sortie {
			sortie[i] = ' '
		}
		cr, rc, err := fn(r.Context(), []byte(req.Entree), sortie, req.CR, req.RC)
		resp := dispatchResponse{Sortie: string(sortie), CR: cr, RC: rc}
		if err != nil {
			resp.Error = err.Error()
			status := http.StatusInternalServerError
			if errors.Is(err, dispatch.ErrCalculatorUnavailable) {
				status = http.StatusUnprocessableEntity
			}
			writeJSON(w, status, resp)
			return
		}
		writeJSON(w, http.StatusOK, resp)
	}
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func buildRouter() http.Handler {
	// EFITAUX2 (TAUDIS store) needs a real store not available here: nil
	// behaves like a failed CALL (CR 24). FMSTAU2 uses the unavailable adapter.
	base := rates.UnavailableBaseTaux{}

	batiCalc, err := bati2018.NewCalculator(nil, base, batiDecoder{})
	if err != nil {
		log.Fatal().Err(err).Msg("bati2018 calculator")
	}
	nonBatiCalc, err := nonbati2018.NewCalculator(nil, base, nonBatiDecoder{})
	if err != nil {
		log.Fatal().Err(err).Msg("nonbati2018 calculator")
	}
	batiDisp, err := dispatch.NewBatiDispatcher(batiCalc)
	if err != nil {
		log.Fatal().Err(err).Msg("bati dispatcher")
	}
	nonBatiDisp, err := dispatch.NewNonBatiDispatcher(nonBatiCalc)
	if err != nil {
		log.Fatal().Err(err).Msg("nonbati dispatcher")
	}

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	r.Post("/api/bati/dispatch", dispatchHandler(func(ctx context.Context, e, s []byte, cr, rc int) (int, int, error) {
		return batiDisp.Dispatch(ctx, e, s, cr, rc)
	}))
	r.Post("/api/nonbati/dispatch", dispatchHandler(func(ctx context.Context, e, s []byte, cr, rc int) (int, int, error) {
		return nonBatiDisp.Dispatch(ctx, e, s, cr, rc)
	}))

	return r
}