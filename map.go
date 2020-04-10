package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go.etcd.io/bbolt"
)

type Config struct {
	Title string   `json:"title"`
	Auths []string `json:"auths"`
}

func (m *Map) getChars(rw http.ResponseWriter, req *http.Request) {
	s := m.getSession(req)
	if s == nil || !s.Auths.Has(AUTH_MAP) {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	chars := []Character{}
	m.chmu.RLock()
	defer m.chmu.RUnlock()
	for _, v := range m.characters {
		chars = append(chars, v)
	}
	json.NewEncoder(rw).Encode(chars)
}

func (m *Map) getMarkers(rw http.ResponseWriter, req *http.Request) {
	s := m.getSession(req)
	if s == nil || !s.Auths.Has(AUTH_MAP) {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	markers := []FrontendMarker{}
	m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("markers"))
		if b == nil {
			return nil
		}
		grid := b.Bucket([]byte("grid"))
		if grid == nil {
			return nil
		}
		grids := tx.Bucket([]byte("grids"))
		if grids == nil {
			return nil
		}
		return grid.ForEach(func(k, v []byte) error {
			m := Marker{}
			json.Unmarshal(v, &m)
			graw := grids.Get([]byte(strconv.Itoa(m.GridID)))
			if graw == nil {
				return nil
			}
			g := GridData{}
			if strings.Contains(m.Name, "BORDER_CAIRN:OURS") {
				m.Image = "gfx/terobjs/mm/cairn"
			}
			if strings.Contains(m.Name, "BORDER_CAIRN,") {
				m.Image = "gfx/terobjs/mm/frendcairn"
			}
			if strings.Contains(m.Name, "SEA_MARK:OURS") {
				m.Image = "gfx/terobjs/mm/seamark"
			}

			if strings.Contains(m.Name, "BORDER_CAIRN:THEIRS,") {
				m.Image = "gfx/terobjs/mm/enemycairn"
			}
			if strings.Contains(m.Name, "SEA_MARK:THEIRS") {
				m.Image = "gfx/terobjs/mm/enemyseamark"
			}

			json.Unmarshal(graw, &g)
			markers = append(markers, FrontendMarker{
				Image:  m.Image,
				Hidden: m.Hidden,
				ID:     m.ID,
				Name:   m.Name,
				Position: Position{
					X: m.Position.X + g.Coord.X*100,
					Y: m.Position.Y + g.Coord.Y*100,
				},
			})
			return nil
		})
	})
	json.NewEncoder(rw).Encode(markers)
}

func (m *Map) config(rw http.ResponseWriter, req *http.Request) {
	s := m.getSession(req)
	if s == nil || !s.Auths.Has(AUTH_MAP) {
		rw.WriteHeader(http.StatusUnauthorized)
		return
	}
	config := Config{
		Auths: s.Auths,
	}
	m.db.View(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("config"))
		if b == nil {
			return nil
		}
		title := b.Get([]byte("title"))
		config.Title = string(title)
		return nil
	})
	json.NewEncoder(rw).Encode(config)
}
