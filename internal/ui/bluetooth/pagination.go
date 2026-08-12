package bluetooth

import (
	"charm.land/bubbles/v2/paginator"
	"charm.land/lipgloss/v2"
	ilovetui "github.com/anotherhadi/ilovetui"
)

func newPaginator() paginator.Model {
	p := paginator.New()
	p.Type = paginator.Dots
	p.ActiveDot = lipgloss.NewStyle().Foreground(ilovetui.S.Primary).Render("•")
	p.InactiveDot = lipgloss.NewStyle().Foreground(ilovetui.S.Subtle).Render("•")
	return p
}

// paginate slices a list of length total into pages of availableRows items,
// deriving the current page from cursor (the globally-indexed selection)
// rather than tracking page state separately. It returns the configured
// paginator and the [start, end) bounds of the visible slice.
func paginate(cursor, total, availableRows int) (p paginator.Model, start, end int) {
	p = newPaginator()
	perPage := availableRows
	if perPage < 1 {
		perPage = 1
	}
	p.PerPage = perPage
	p.SetTotalPages(total)
	if p.TotalPages > 0 {
		p.Page = cursor / perPage
		if p.Page >= p.TotalPages {
			p.Page = p.TotalPages - 1
		}
	}
	start, end = p.GetSliceBounds(total)
	return p, start, end
}

func renderPaginatorDots(p paginator.Model) string {
	if p.TotalPages <= 1 {
		return ""
	}
	return mutedStyle().Render(p.View())
}
