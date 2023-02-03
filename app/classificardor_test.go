package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCanBeApproved(t *testing.T) {

	// Arrange
	sabado := []int{6}
	domingo := []int{7}
	ambos := []int{6, 7}

	scenarios := []struct {
		name    string
		current []int
		already []int
		expect  bool
	}{
		{name: "primeira aprovação no sábado", current: sabado, already: nil, expect: true},
		{name: "primeira aprovação no domingo", current: domingo, already: nil, expect: true},
		{name: "primeira aprovação em ambos", current: ambos, already: nil, expect: true},
		{name: "aprovação no sábado quando já aprovado no domingo", current: sabado, already: domingo, expect: true},
		{name: "aprovação no domingo quando já aprovado no sábado", current: domingo, already: sabado, expect: true},
		{name: "aprovação no sábado quando já aprovado no sábado", current: sabado, already: sabado, expect: false},
		{name: "aprovação no domingo quando já aprovado no domingo", current: domingo, already: domingo, expect: false},
		{name: "aprovação em ambos quando já aprovado no sabado", current: ambos, already: sabado, expect: false},
		{name: "aprovação em ambos quando já aprovado no domingo", current: ambos, already: domingo, expect: false},
		{name: "aprovação no sábado quando já aprovado em ambos", current: sabado, already: ambos, expect: false},
		{name: "aprovação no domingo quando já aprovado em ambos", current: domingo, already: ambos, expect: false},
		{name: "aprovação em ambos quando já aprovado em ambos", current: ambos, already: ambos, expect: false},
	}

	for _, s := range scenarios {

		// Act
		got := canBeApproved(s.current, s.already)

		// Assert
		assert.Equal(t, got, s.expect, fmt.Sprintf("%s deveria retornar %t", s.name, s.expect))

	}

}
