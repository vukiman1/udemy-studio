import type { CSSProperties, ReactElement } from 'react'
import type { CourseSummary } from '../../api/types'

// The five tile hues of the design's "Course avatar"; a course keeps the same one on every visit.
const TILE_HUES = ['#8b6cff', '#4f8cff', '#e26ba8', '#2fc4b2', '#f2a93b'] as const
const INITIALS = 2

type TileStyle = CSSProperties & Record<'--tile', string>

interface CourseThumbProps {
  readonly course: CourseSummary
  readonly className: string
}

export function CourseThumb({ course, className }: CourseThumbProps): ReactElement {
  if (course.coverUrl !== undefined)
    return (
      <img className={`${className} object-cover`} src={course.coverUrl} alt="" loading="lazy" />
    )
  const style: TileStyle = { '--tile': tileHue(course.id) }
  return (
    <div
      className={`${className} grid place-items-center border border-[color-mix(in_srgb,var(--tile)_35%,transparent)] bg-[color-mix(in_srgb,var(--tile)_20%,transparent)] text-2xl font-bold text-[color-mix(in_srgb,var(--tile)_75%,black)] dark:text-[color-mix(in_srgb,var(--tile)_60%,white)]`}
      style={style}
      aria-hidden="true"
    >
      {initials(course.title)}
    </div>
  )
}

function tileHue(courseId: string): string {
  let sum = 0
  for (let position = 0; position < courseId.length; position++)
    sum += courseId.charCodeAt(position)
  return TILE_HUES[sum % TILE_HUES.length] ?? TILE_HUES[0]
}

// "React & TypeScript" → "RT": the first letter of each of the first two words, skipping symbols.
function initials(title: string): string {
  return title
    .split(/[^\p{L}\p{N}]+/u)
    .filter((word) => word !== '')
    .slice(0, INITIALS)
    .map((word) => word.slice(0, 1).toUpperCase())
    .join('')
}
