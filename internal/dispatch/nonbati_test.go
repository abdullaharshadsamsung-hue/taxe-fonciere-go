package dispatch

import (
	"bytes"
	"context"
	"errors"
	"testing"

	"migrated-app/internal/calc/nonbati2018"
)

type mockNonBati struct {
	calls     int
	gotEntree []byte
	gotParm   string
	out       nonbati2018.Output
	err       error
}

func (m *mockNonBati) Calculate(ctx context.Context, comnonbam []byte, parm string) (nonbati2018.Output, error) {
	m.calls++
	m.gotEntree = append([]byte(nil), comnonbam...)
	m.gotParm = parm
	return m.out, m.err
}

func nbMakeEntree(year string) []byte {
	b := bytes.Repeat([]byte{'X'}, RecordLen)
	copy(b[1:5], year)
	return b
}

func nbMakeSortie() []byte {
	return bytes.Repeat([]byte{'S'}, RecordLen)
}

func TestNewNonBatiDispatcherNil(t *testing.T) {
	d, err := NewNonBatiDispatcher(nil)
	if err == nil || d != nil {
		t.Fatalf("expected error and nil dispatcher, got %v %v", d, err)
	}
}

func TestNonBatiYearProgram(t *testing.T) {
	tests := []struct {
		year string
		want string
		ok   bool
	}{
		{"2010", "EFITA3N0", true},
		{"2011", "EFITA3N1", true},
		{"2012", "EFITA3N2", true},
		{"2013", "EFITA3N3", true},
		{"2014", "EFITA3N4", true},
		{"2015", "EFITA3N5", true},
		{"2016", "EFITA3N6", true},
		{"2017", "EFITA3N7", true},
		{"2018", "EFITA3N8", true},
		{"2009", "", false},
		{"2019", "", false},
		{"    ", "", false},
		{"ABCD", "", false},
		{"18", "", false},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.year, func(t *testing.T) {
			got, ok := NonBatiYearProgram(tt.year)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("NonBatiYearProgram(%q) = %q,%v want %q,%v", tt.year, got, ok, tt.want, tt.ok)
			}
		})
	}
}

func TestNonBatiDispatch2018(t *testing.T) {
	tests := []struct {
		name     string
		inCR     int
		wantParm string
	}{
		{"croisic 99", 99, nonbati2018.ParmTaudis},
		{"cr 00", 0, " "},
		{"cr 12", 12, " "},
		{"cr 98", 98, " "},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			var out nonbati2018.Output
			copy(out.Retour[:], bytes.Repeat([]byte{'R'}, RecordLen))
			out.CR = 12
			out.RC = 3
			m := &mockNonBati{out: out}
			d, err := NewNonBatiDispatcher(m)
			if err != nil {
				t.Fatal(err)
			}
			entree := nbMakeEntree("2018")
			orig := append([]byte(nil), entree...)
			sortie := nbMakeSortie()
			cr, rc, err := d.Dispatch(context.Background(), entree, sortie, tt.inCR, 7)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if m.calls != 1 {
				t.Fatalf("calls=%d want 1", m.calls)
			}
			if m.gotParm != tt.wantParm {
				t.Fatalf("parm=%q want %q", m.gotParm, tt.wantParm)
			}
			if !bytes.Equal(m.gotEntree, orig) || !bytes.Equal(entree, orig) {
				t.Fatal("entree modified or not passed intact")
			}
			if cr != 12 || rc != 3 {
				t.Fatalf("cr,rc=%d,%d want 12,3", cr, rc)
			}
			if !bytes.Equal(sortie, out.Retour[:]) {
				t.Fatal("sortie not copied from calculator output")
			}
		})
	}
}

func TestNonBatiDispatchUnportedYears(t *testing.T) {
	for _, year := range []string{"2010", "2011", "2012", "2013", "2014", "2015", "2016", "2017"} {
		for _, inCR := range []int{0, 99} {
			year, inCR := year, inCR
			t.Run(year, func(t *testing.T) {
				m := &mockNonBati{}
				d, _ := NewNonBatiDispatcher(m)
				sortie := nbMakeSortie()
				orig := nbMakeSortie()
				cr, rc, err := d.Dispatch(context.Background(), nbMakeEntree(year), sortie, inCR, 5)
				if !errors.Is(err, ErrCalculatorUnavailable) {
					t.Fatalf("err=%v want ErrCalculatorUnavailable", err)
				}
				if cr != 0 || rc != 5 {
					t.Fatalf("cr,rc=%d,%d want 0,5", cr, rc)
				}
				if m.calls != 0 {
					t.Fatal("calculator should not be called")
				}
				if !bytes.Equal(sortie, orig) {
					t.Fatal("sortie modified")
				}
			})
		}
	}
}

func TestNonBatiDispatchUnknownYear(t *testing.T) {
	tests := []struct {
		name   string
		entree []byte
		inCR   int
	}{
		{"2009", nbMakeEntree("2009"), 0},
		{"2019", nbMakeEntree("2019"), 99},
		{"spaces", nbMakeEntree("    "), 99},
		{"nonnumeric", nbMakeEntree("ABCD"), 1},
		{"empty entree", []byte{}, 99},
		{"short entree", []byte(" 20"), 0},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := &mockNonBati{}
			d, _ := NewNonBatiDispatcher(m)
			sortie := nbMakeSortie()
			orig := nbMakeSortie()
			cr, rc, err := d.Dispatch(context.Background(), tt.entree, sortie, tt.inCR, 42)
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if cr != 0 || rc != 42 {
				t.Fatalf("cr,rc=%d,%d want 0,42", cr, rc)
			}
			if m.calls != 0 {
				t.Fatal("calculator should not be called")
			}
			if !bytes.Equal(sortie, orig) {
				t.Fatal("sortie modified")
			}
		})
	}
}

func TestNonBatiDispatchCalculatorError(t *testing.T) {
	sentinel := errors.New("boom")
	m := &mockNonBati{err: sentinel}
	d, _ := NewNonBatiDispatcher(m)
	sortie := nbMakeSortie()
	orig := nbMakeSortie()
	cr, rc, err := d.Dispatch(context.Background(), nbMakeEntree("2018"), sortie, 99, 8)
	if !errors.Is(err, sentinel) {
		t.Fatalf("err=%v want wrapped sentinel", err)
	}
	if cr != 0 || rc != 8 {
		t.Fatalf("cr,rc=%d,%d want 0,8", cr, rc)
	}
	if m.calls != 1 {
		t.Fatalf("calls=%d", m.calls)
	}
	if !bytes.Equal(sortie, orig) {
		t.Fatal("sortie modified on error")
	}
}

func TestNonBatiDispatchShortSortie(t *testing.T) {
	tests := []struct {
		name string
		len  int
	}{
		{"nil", 0},
		{"599", RecordLen - 1},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := &mockNonBati{}
			d, _ := NewNonBatiDispatcher(m)
			cr, rc, err := d.Dispatch(context.Background(), nbMakeEntree("2018"), make([]byte, tt.len), 99, 4)
			if err == nil {
				t.Fatal("expected error")
			}
			if cr != 99 || rc != 4 {
				t.Fatalf("cr,rc=%d,%d want 99,4", cr, rc)
			}
			if m.calls != 0 {
				t.Fatal("calculator should not be called")
			}
		})
	}
}