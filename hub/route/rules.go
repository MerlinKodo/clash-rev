package route

import (
	"net/http"
	"strconv"

	"github.com/MerlinKodo/clash-rev/constant"

	"github.com/MerlinKodo/clash-rev/tunnel"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

func ruleRouter() http.Handler {
	r := chi.NewRouter()
	r.Get("/", getRules)
	return r
}

type Rule struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
	Proxy   string `json:"proxy"`
	Size    int    `json:"size"`
}

func getRules(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	sizeStr := r.URL.Query().Get("size")

	var page, size, start, end int
	var errPage, errSize error
	paginate := true

	if pageStr != "" && sizeStr != "" {
		page, errPage = strconv.Atoi(pageStr)
		size, errSize = strconv.Atoi(sizeStr)

		if errPage != nil || errSize != nil || page <= 0 || size <= 0 {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, render.M{"error": "Invalid page or size number"})
			return
		}

		start = (page - 1) * size
		end = start + size
	} else {
		paginate = false
	}

	rawRules := tunnel.Rules()
	totalRules := len(rawRules)

	if !paginate {
		start = 0
		end = totalRules
	} else if start >= totalRules {
		start = totalRules
		end = totalRules
	} else if end > totalRules {
		end = totalRules
	}

	rules := []Rule{}
	for _, rule := range rawRules[start:end] {
		r := Rule{
			Type:    rule.RuleType().String(),
			Payload: rule.Payload(),
			Proxy:   rule.Adapter(),
			Size:    -1,
		}
		if rule.RuleType() == constant.GEOIP || rule.RuleType() == constant.GEOSITE {
			r.Size = rule.(constant.RuleGroup).GetRecodeSize()
		}
		rules = append(rules, r)
	}

	render.JSON(w, r, render.M{
		"rules": rules,
	})
}
