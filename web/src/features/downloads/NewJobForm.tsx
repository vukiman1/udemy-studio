import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useState, type ReactElement } from 'react'
import { createJob } from '../../api/download-queries'
import { QUERY_KEYS } from '../../api/queries'
import type { Job } from '../../api/types'
import { upsertJob } from './job-list'

const COURSE_URL_PLACEHOLDER = 'https://<portal>.udemy.com/course/<tên-khoá>/'

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
      className="grid gap-3 rounded-xl border border-line bg-surface p-4"
      onSubmit={(event) => {
        event.preventDefault()
        submitForm()
      }}
    >
      <h2 className="font-semibold">Tải khoá học mới</h2>
      <label className="grid gap-1 text-[0.9rem]">
        Link khoá học
        <input
          className="rounded-md border border-line bg-page px-2.5 py-1.5"
          type="url"
          required
          placeholder={COURSE_URL_PLACEHOLDER}
          value={courseUrl}
          onChange={(event) => {
            setCourseUrl(event.target.value)
          }}
        />
      </label>
      <label className="grid gap-1 text-[0.9rem]">
        Chỉ tải chapter (để trống để tải hết), ví dụ 1,3-5
        <input
          className="rounded-md border border-line bg-page px-2.5 py-1.5"
          inputMode="numeric"
          value={chapters}
          onChange={(event) => {
            setChapters(event.target.value)
          }}
        />
      </label>
      <div className="flex items-center gap-3">
        <button type="submit" disabled={submit.isPending}>
          Bắt đầu tải
        </button>
        {submit.isError && <p className="text-danger">{submit.error.message}</p>}
      </div>
    </form>
  )
}
