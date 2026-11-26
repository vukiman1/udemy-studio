import type { PublishPlan } from '../../api/types'

export function initialSelection(plan: PublishPlan, courseId: string): ReadonlySet<string> {
  const isThisCourseOnR2 = plan.published?.courseId === courseId
  const chapters = isThisCourseOnR2
    ? plan.chapters.filter((chapter) => chapter.published)
    : plan.chapters
  return new Set(chapters.map((chapter) => chapter.dir))
}

export function selectedBytes(plan: PublishPlan, selected: ReadonlySet<string>): number {
  return plan.chapters
    .filter((chapter) => selected.has(chapter.dir))
    .reduce((total, chapter) => total + chapter.sizeBytes, plan.sharedBytes)
}
