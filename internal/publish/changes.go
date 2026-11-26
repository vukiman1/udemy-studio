package publish

import (
	"slices"
	"strings"
)

func (plan Plan) SharedChanged() bool {
	if plan.hasUnsentFile("") {
		return true
	}
	local := make(map[string]bool, len(plan.Files))
	chapters := map[string]bool{}
	for _, file := range plan.Files {
		local[ObjectKey(plan.Course.Dir, file.Path)] = true
		chapters[chapterOf(file.Path)] = true
	}
	for key := range plan.Remote {
		chapter := chapterOf(strings.TrimPrefix(key, ObjectKey(plan.Course.Dir, "")+"/"))
		if !local[key] && (chapter == "" || !chapters[chapter]) {
			return true
		}
	}
	return false
}

func (plan Plan) hasChanges(chapter string) bool {
	if plan.hasUnsentFile(chapter) {
		return true
	}
	prefix := ObjectKey(plan.Course.Dir, chapter) + "/"
	for key := range plan.Remote {
		isLocal := slices.ContainsFunc(plan.Files, func(file courseFile) bool { return ObjectKey(plan.Course.Dir, file.Path) == key })
		if strings.HasPrefix(key, prefix) && !isLocal {
			return true
		}
	}
	return false
}

func (plan Plan) hasUnsentFile(chapter string) bool {
	return slices.ContainsFunc(plan.Files, func(file courseFile) bool {
		if chapterOf(file.Path) != chapter {
			return false
		}
		object, found := plan.Remote[ObjectKey(plan.Course.Dir, file.Path)]
		return !found || !isUpToDate(object, file)
	})
}
