package plan

import "sort"

type DesiredInstance struct {
	InstanceID    string
	TemplateName  string
	ContainerName string
	Enabled       bool
}

func ComputeDiff(desired []DesiredInstance, running []string) []Change {
	runningSet := make(map[string]bool)
	for _, name := range running {
		runningSet[name] = true
	}

	desiredMap := make(map[string]DesiredInstance)
	for _, d := range desired {
		desiredMap[d.ContainerName] = d
	}

	var changes []Change

	for _, d := range desired {
		isRunning := runningSet[d.ContainerName]

		if d.Enabled && !isRunning {
			changes = append(changes, Change{
				Action:        ActionAdd,
				InstanceID:    d.InstanceID,
				TemplateName:  d.TemplateName,
				ContainerName: d.ContainerName,
				Fields:        nil,
			})
		} else if !d.Enabled && isRunning {
			changes = append(changes, Change{
				Action:        ActionModify,
				InstanceID:    d.InstanceID,
				TemplateName:  d.TemplateName,
				ContainerName: d.ContainerName,
				Fields: map[string]FieldChange{
					"enabled": {
						Old:    true,
						New:    false,
						Source: "user",
					},
				},
			})
		}
	}

	for name := range runningSet {
		if _, exists := desiredMap[name]; !exists {
			changes = append(changes, Change{
				Action:        ActionRemove,
				InstanceID:    "",
				TemplateName:  "",
				ContainerName: name,
				Fields:        nil,
			})
		}
	}

	sort.Slice(changes, func(i, j int) bool {
		order := map[Action]int{
			ActionRemove: 0,
			ActionModify: 1,
			ActionAdd:    2,
		}
		return order[changes[i].Action] < order[changes[j].Action]
	})

	return changes
}
