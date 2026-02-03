package container

import "testing"

func TestBuildLabels(t *testing.T) {
	labels := BuildLabels("my-stack", "postgres-1", "42")

	expectedKeys := []string{
		LabelManagedBy,
		LabelVersion,
		LabelStackID,
		LabelInstanceID,
		LabelTemplateServiceID,
	}

	for _, key := range expectedKeys {
		if _, ok := labels[key]; !ok {
			t.Errorf("BuildLabels missing expected key: %s", key)
		}
	}

	if labels[LabelManagedBy] != ManagedByValue {
		t.Errorf("Expected managed_by=%s, got %s", ManagedByValue, labels[LabelManagedBy])
	}
	if labels[LabelStackID] != "my-stack" {
		t.Errorf("Expected stack_id=my-stack, got %s", labels[LabelStackID])
	}
	if labels[LabelInstanceID] != "postgres-1" {
		t.Errorf("Expected instance_id=postgres-1, got %s", labels[LabelInstanceID])
	}
	if labels[LabelTemplateServiceID] != "42" {
		t.Errorf("Expected template_service_id=42, got %s", labels[LabelTemplateServiceID])
	}
}

func TestBuildLabelsPartial(t *testing.T) {
	labels := BuildLabels("stack1", "", "")

	if labels[LabelStackID] != "stack1" {
		t.Errorf("Expected stack_id=stack1, got %s", labels[LabelStackID])
	}
	if _, ok := labels[LabelInstanceID]; ok {
		t.Error("Expected no instance_id label when empty")
	}
	if _, ok := labels[LabelTemplateServiceID]; ok {
		t.Error("Expected no template_service_id label when empty")
	}
}

func TestContainerName(t *testing.T) {
	tests := []struct {
		stackID    string
		instanceID string
		expected   string
	}{
		{"my-stack", "postgres-1", "devarch-my-stack-postgres-1"},
		{"prod", "api", "devarch-prod-api"},
		{"a", "b", "devarch-a-b"},
	}

	for _, tc := range tests {
		result := ContainerName(tc.stackID, tc.instanceID)
		if result != tc.expected {
			t.Errorf("ContainerName(%s, %s) = %s, expected %s",
				tc.stackID, tc.instanceID, result, tc.expected)
		}
	}
}

func TestNetworkName(t *testing.T) {
	tests := []struct {
		stackID  string
		expected string
	}{
		{"my-stack", "devarch-my-stack-net"},
		{"prod", "devarch-prod-net"},
		{"a", "devarch-a-net"},
	}

	for _, tc := range tests {
		result := NetworkName(tc.stackID)
		if result != tc.expected {
			t.Errorf("NetworkName(%s) = %s, expected %s",
				tc.stackID, result, tc.expected)
		}
	}
}

func TestIsDevArchManaged(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		expected bool
	}{
		{
			"managed container",
			map[string]string{LabelManagedBy: ManagedByValue},
			true,
		},
		{
			"unmanaged container",
			map[string]string{"foo": "bar"},
			false,
		},
		{
			"wrong managed_by value",
			map[string]string{LabelManagedBy: "other"},
			false,
		},
		{
			"empty labels",
			map[string]string{},
			false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := IsDevArchManaged(tc.labels)
			if result != tc.expected {
				t.Errorf("IsDevArchManaged(%v) = %v, expected %v",
					tc.labels, result, tc.expected)
			}
		})
	}
}
