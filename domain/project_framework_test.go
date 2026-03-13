package domain

import "testing"

func TestValidateFrameworkConfig_NilIsValid(t *testing.T) {
	if err := ValidateFrameworkConfig(nil); err != nil {
		t.Errorf("nil FrameworkConfig should be valid, got: %v", err)
	}
}

func TestValidateFrameworkConfig_ValidStandalone(t *testing.T) {
	fc := &FrameworkConfig{Version: "2026.03.1", Mode: "standalone"}
	if err := ValidateFrameworkConfig(fc); err != nil {
		t.Errorf("valid standalone config should pass, got: %v", err)
	}
}

func TestValidateFrameworkConfig_ValidThin(t *testing.T) {
	fc := &FrameworkConfig{Version: "2026.03.1", Mode: "thin"}
	if err := ValidateFrameworkConfig(fc); err != nil {
		t.Errorf("valid thin config should pass, got: %v", err)
	}
}

func TestValidateFrameworkConfig_ModeDefaultsToStandalone(t *testing.T) {
	fc := &FrameworkConfig{Version: "2026.03.1"}
	if err := ValidateFrameworkConfig(fc); err != nil {
		t.Errorf("empty mode should be valid (defaults to standalone), got: %v", err)
	}
	if fc.DeploymentModeOrDefault() != ModeStandalone {
		t.Errorf("expected ModeStandalone, got %v", fc.DeploymentModeOrDefault())
	}
}

func TestValidateFrameworkConfig_MissingVersion(t *testing.T) {
	fc := &FrameworkConfig{Mode: "standalone"}
	if err := ValidateFrameworkConfig(fc); err == nil {
		t.Error("expected error for missing version")
	}
}

func TestValidateFrameworkConfig_InvalidCalVer(t *testing.T) {
	cases := []string{"1.2.3", "2026.3.1", "2026.03", "v2026.03.1", "abc"}
	for _, v := range cases {
		fc := &FrameworkConfig{Version: v}
		if err := ValidateFrameworkConfig(fc); err == nil {
			t.Errorf("expected error for invalid version %q", v)
		}
	}
}

func TestValidateFrameworkConfig_InvalidMode(t *testing.T) {
	fc := &FrameworkConfig{Version: "2026.03.1", Mode: "hybrid"}
	if err := ValidateFrameworkConfig(fc); err == nil {
		t.Error("expected error for invalid mode 'hybrid'")
	}
}

func TestDeploymentModeOrDefault_NilConfig(t *testing.T) {
	var fc *FrameworkConfig
	if fc.DeploymentModeOrDefault() != ModeStandalone {
		t.Error("nil FrameworkConfig should default to standalone")
	}
}
