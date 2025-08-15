package customtype

import (
	"strings"
	"time"
)

type CustomTime struct {
    time.Time
}

const ctLayout = "15:04"

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
    s := strings.Trim(string(b), `"`)
    t, err := time.Parse(ctLayout, s)
    if err != nil {
        return err
    }
    ct.Time = t
    return nil
}

func (ct CustomTime) MarshalJSON() ([]byte, error) {
    return []byte(`"` + ct.Format(ctLayout) + `"`), nil
}

