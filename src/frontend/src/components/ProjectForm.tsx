import { useQueryClient } from '@tanstack/react-query';
import { useState } from 'react';
import { createProject, updateProject, type ProjectDto } from '../api/projects';

const inputClass = "w-full border border-gray-200 rounded px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500"
const labelClass = "block text-sm font-medium text-gray-700 mb-1"

interface Props {
    workspaceId: string
    project?: ProjectDto
    onClose: () => void
}

export default function ProjectForm({ workspaceId, project, onClose }: Props) {
    const queryClient = useQueryClient()
    const isEdit = !!project

    const [name, setName] = useState(project?.name ?? '')
    const [imageUrl, setImageUrl] = useState(project?.image_url ?? '')
    const [error, setError] = useState('')
    const [loading, setLoading] = useState(false)

    async function handleSubmit(e: React.FormEvent) {
        e.preventDefault()
        setError('')
        setLoading(true)
        try {
            if (isEdit) {
                await updateProject(workspaceId, project.id, { name, image_url: imageUrl || undefined })
            } else {
                await createProject(workspaceId, { name, image_url: imageUrl || undefined })
            }
            await queryClient.invalidateQueries({ queryKey: ['projects', workspaceId] })
            onClose()
        } catch {
            setError('Something went wrong. Please try again')
        } finally {
            setLoading(false)
        }
    }

    return (
        <form onSubmit={handleSubmit} className="flex flex-col gap-4">
            {error && <p className="text-red-500 text-sm">{error}</p>}

            <div>
                <label className={labelClass}>Name <span className="text-red-500">*</span></label>
                <input
                    type="text"
                    value={name}
                    onChange={e => setName(e.target.value)}
                    className={inputClass}
                    required
                />
            </div>

            <div>
                <label className={labelClass}>Image URL</label>
                <input
                    type="text"
                    value={imageUrl}
                    onChange={e => setImageUrl(e.target.value)}
                    className={inputClass}
                    placeholder="https://..."
                />
            </div>

            <div className="flex justify-end gap-3 pt-2">
                <button type="button" onClick={onClose} className="px-4 py-2 text-sm text-gray-600
  hover:text-gray-800">
                    Cancel
                </button>
                <button
                    type="submit"
                    disabled={loading}
                    className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700
  disabled:opacity-50"
                >
                    {loading ? 'Saving...' : isEdit ? 'Save changes' : 'Create project'}
                </button>
            </div>
        </form>
    )
}