package domain

import (
	"errors"
	"testing"
)

func TestParseTipoDocumento(t *testing.T) {
	casos := []struct {
		entrada string
		quer    TipoDocumento
		erro    bool
	}{
		{"LAUDO", TipoLaudo, false},
		{" exame_previo ", TipoExamePrevio, false},
		{"Receita", TipoReceita, false},
		{"ATESTADO", TipoAtestado, false},
		{"", "", true},
		{"RAIO_X", "", true},
	}
	for _, c := range casos {
		obtido, err := ParseTipoDocumento(c.entrada)
		if c.erro {
			if !errors.Is(err, ErrTipoInvalido) {
				t.Errorf("ParseTipoDocumento(%q): esperava ErrTipoInvalido, veio %v", c.entrada, err)
			}
			continue
		}
		if err != nil || obtido != c.quer {
			t.Errorf("ParseTipoDocumento(%q) = %q, %v; quer %q", c.entrada, obtido, err, c.quer)
		}
	}
}
