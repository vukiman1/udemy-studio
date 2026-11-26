import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { Download } from 'lucide-react'
import { createJob } from '../../api/download-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { Job } from '../../api/types'
import { Button, ButtonSize, ButtonVariant } from '../../ui/Button'
import { TEXT_FIELD } from '../../ui/text-field'
import { upsertJob } from './job-list'

const COURSE_URL_PLACEHOLDER = 'https://<portal>.udemy.com/course/<course-slug>/'

export function NewJobForm(): ReactElement {
  const queryClient = useQueryClient()
  const [courseUrl, setCourseUrl] = useState('')
  const [chapters, setChapters] = useState('')
  const submit = useMutation({
    mutationFn: createJob,
    onSuccess: (job) => {
      queryClient.setQueryData<Job[]>(QUERY_KEYS.jobs, (current) => upsertJob(current ?? [], job))
      setCourseUrl('')
      setChapters('')
    },
  })

  const submitForm = (): void => {
    const trimmedChapters = chapters.trim()
    submit.mutate({
      courseUrl: courseUrl.trim(),
      ...(trimmedChapters === '' ? {} : { chapters: trimmedChapters }),
    })
  }

  return (
    <form
      className="grid gap-[18px] rounded-[20px] bg-card p-6"
      onSubmit={(event) => {
        event.preventDefault()
        submitForm()
      }}
    >
      <h2 className="text-base font-semibold">New download</h2>
      <div className="flex flex-wrap items-end gap-3">
        <label className="grid min-w-64 flex-1 gap-2 text-sm font-semibold">
          Course link
          <input
            className={`${TEXT_FIELD} font-normal`}
            type="url"
            required
            placeholder={COURSE_URL_PLACEHOLDER}
            value={courseUrl}
            onChange={(event) => {
              setCourseUrl(event.target.value)
            }}
          />
        </label>
        <label className="grid w-full gap-2 text-sm font-semibold sm:w-[280px]">
          Chapters
          <input
            className={`${TEXT_FIELD} font-normal`}
            placeholder="Empty for all · e.g. 1,3-5"
            value={chapters}
            onChange={(event) => {
              setChapters(event.target.value)
            }}
          />
        </label>
        <Button
          type="submit"
          variant={ButtonVariant.Primary}
          size={ButtonSize.Large}
          icon={Download}
          disabled={submit.isPending}
        >
          Download
        </Button>
      </div>
      {submit.isError && (
        <p className="text-[13px] text-danger" role="alert">
          {submit.error.message}
        </p>
      )}
    </form>
  )
}
