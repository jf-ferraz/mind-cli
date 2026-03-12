package tui

import (
	"fmt"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jf-ferraz/mind-cli/domain"
)

func TestStatusView_InitialState(t *testing.T) {
	v := NewStatusView()
	if v.viewState != ViewLoading {
		t.Errorf("initial viewState = %d, want ViewLoading (%d)", v.viewState, ViewLoading)
	}
}

func TestStatusView_HealthLoaded(t *testing.T) {
	v := NewStatusView()
	health := &domain.ProjectHealth{
		Zones: map[domain.Zone]domain.ZoneHealth{
			domain.ZoneSpec: {Zone: domain.ZoneSpec, Total: 5, Complete: 4},
		},
	}

	v, _ = v.Update(healthLoadedMsg{health: health})

	if v.viewState != ViewReady {
		t.Errorf("viewState = %d, want ViewReady (%d)", v.viewState, ViewReady)
	}
	if v.health == nil {
		t.Fatal("health is nil after healthLoadedMsg")
	}
	if v.health.Zones[domain.ZoneSpec].Complete != 4 {
		t.Errorf("spec complete = %d, want 4", v.health.Zones[domain.ZoneSpec].Complete)
	}
}

func TestStatusView_HealthNil(t *testing.T) {
	v := NewStatusView()
	v, _ = v.Update(healthLoadedMsg{health: nil})

	if v.viewState != ViewEmpty {
		t.Errorf("viewState = %d, want ViewEmpty (%d)", v.viewState, ViewEmpty)
	}
}

func TestStatusView_HealthError(t *testing.T) {
	v := NewStatusView()
	v, _ = v.Update(healthErrorMsg{err: fmt.Errorf("test error")})

	if v.viewState != ViewError {
		t.Errorf("viewState = %d, want ViewError (%d)", v.viewState, ViewError)
	}
	if v.errMsg != "test error" {
		t.Errorf("errMsg = %q, want 'test error'", v.errMsg)
	}
}

func TestStatusView_WindowResize(t *testing.T) {
	v := NewStatusView()
	v, _ = v.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	if v.width != 120 {
		t.Errorf("width = %d, want 120", v.width)
	}
	if v.height != 40 {
		t.Errorf("height = %d, want 40", v.height)
	}
}

func TestStatusView_ViewLoading(t *testing.T) {
	v := NewStatusView()
	v.width = 80
	v.height = 24
	output := v.View()
	if !strings.Contains(output, "Loading") {
		t.Error("expected loading message in View()")
	}
}

func TestStatusView_ViewError(t *testing.T) {
	v := NewStatusView()
	v.viewState = ViewError
	v.errMsg = "something failed"
	v.width = 80
	v.height = 24
	output := v.View()
	if !strings.Contains(output, "something failed") {
		t.Error("expected error message in View()")
	}
	if !strings.Contains(output, "retry") {
		t.Error("expected retry hint in View()")
	}
}

func TestStatusView_ViewReady_TwoColumn(t *testing.T) {
	v := NewStatusView()
	v.viewState = ViewReady
	v.width = 100
	v.height = 40
	v.health = &domain.ProjectHealth{
		Zones: map[domain.Zone]domain.ZoneHealth{
			domain.ZoneSpec:       {Zone: domain.ZoneSpec, Total: 5, Complete: 3},
			domain.ZoneBlueprints: {Zone: domain.ZoneBlueprints, Total: 2, Complete: 2},
			domain.ZoneState:      {Zone: domain.ZoneState, Total: 3, Complete: 3},
		},
		Warnings:   []string{"Missing brief"},
		Suggestions: []string{"Run mind init"},
	}

	output := v.View()
	if !strings.Contains(output, "Documentation Health") {
		t.Error("expected 'Documentation Health' heading")
	}
	if !strings.Contains(output, "Workflow") {
		t.Error("expected 'Workflow' section")
	}
	if !strings.Contains(output, "Quick Actions") {
		t.Error("expected 'Quick Actions' section")
	}
}
