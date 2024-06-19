package main

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/totalys/meimei-classificador/pkg/domain"
)

func TestCanBeApproved(t *testing.T) {

	// Arrange
	sabado := []int{6}
	domingo := []int{7}
	ambos := []int{6, 7}
	student := domain.Student{
		Entrevistas: map[string]float64{
			"Course A": 10,
		},
	}

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
		got := canBeApproved(domain.CourseConfig{
			Days: s.current,
		}, &student, s.already, "Course A")

		// Assert
		assert.Equal(t, s.expect, got, fmt.Sprintf("%s deveria retornar %t", s.name, s.expect))

	}

}

// Regra está desligada
func TestCanBeApprovedBySenaiTwice(t *testing.T) {

	// Arrange
	student := domain.Student{
		Approved: []domain.Approved{
			{
				IsSenai: true,
				Course:  "Course A",
				Days:    []int{6},
			}},
		Entrevistas: map[string]float64{
			"Course A": 10,
		},
	}

	scenarios := []struct {
		name    string
		student domain.Student
		course  domain.CourseConfig
		expect  bool
	}{
		{
			name:    "Curso atual não é do Senai",
			student: student,
			course:  domain.CourseConfig{IsSenai: false},
			expect:  true,
		},
		{

			name:    "Curso atual é do Senai",
			student: student,
			course:  domain.CourseConfig{IsSenai: true},
			expect:  false,
		},
	}

	for _, s := range scenarios {

		// Act
		got := canBeApproved(s.course, &s.student, []int{}, "Course A")

		// Assert
		assert.Equal(t, s.expect, got, fmt.Sprintf("%s deveria retornar %t", s.student.Name, s.expect))

	}

}
