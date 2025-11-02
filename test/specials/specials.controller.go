package sanity_test

import (
	"context"
	"time"

	"github.com/gopher-fleece/runtime"
)

// @Tag(Specials Controller Tag)
// @Route(/test/specials)
// @Description Specials Controller
type SpecialsController struct {
	runtime.GleeceController // Embedding the GleeceController to inherit its methods
}

// @Method(POST)
// @Route(/accepts-any)
// @Body(body)
func (ec *SpecialsController) ReceivesAny(body any) error {
	return nil
}

// @Method(POST)
// @Route(/returns-any)
func (ec *SpecialsController) ReturnsAny() (any, error) {
	return nil, nil
}

// @Method(POST)
// @Route(/accepts-time)
// @Body(body)
func (ec *SpecialsController) ReceivesTime(body time.Time) error {
	return nil
}

// @Method(POST)
// @Route(/returns-time)
func (ec *SpecialsController) ReturnsTime() (time.Time, error) {
	return time.Now(), nil
}

// @Method(POST)
// @Route(/accepts-time)
// @Body(body)
func (ec *SpecialsController) ReceivesContext(body context.Context) error {
	return nil
}

// @Method(POST)
// @Route(/returns-time)
func (ec *SpecialsController) ReturnsContext() (context.Context, error) {
	return nil, nil
}
