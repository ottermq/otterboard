import { useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { deleteProject, type ProjectDto } from '../../api/projects'
import Modal from '../../components/Modal'
import ProjectForm from '../../components/ProjectForm'
import { useProjects } from '../../hooks/useProjecs'

export default function ProjectsPage() {
    const { workspaceId } = useParams<{ workspaceId: string }>()
    const queryClient = useQueryClient()
    const { data: projects, isLoading, isError } = useProjects(workspaceId!)

    const [showCreate, setShowCreate] = useState(false)
    const [editProject, setEditProject] = useState<ProjectDto | null>(null)
    const [deletingId, setDeletingId] = useState<string | null>(null)

    async function handleDelete(projectId: string) {
        if (!confirm('Delete this project? This cannot be undone.')) return
        setDeletingId(projectId)
        try {
            await deleteProject(workspaceId!, projectId)
            await queryClient.invalidateQueries({ queryKey: ['projects', workspaceId] })
        } finally {
            setDeletingId(null)
        }
    }

    if (isLoading) return <div className="p-8 text-gray-500">Loading...</div>
    if (isError) return <div className="p-8 text-red-500">Failed to load projects.</div>

    return (
        <div className="p-8 max-w-4xl mx-auto">
            <div className="flex items-center justify-between mb-6">
                <h1 className="text-2xl font-semibold">Projects</h1>
                <button
                    onClick={() => setShowCreate(true)}
                    className="px-4 py-2 text-sm bg-blue-600 text-white rounded hover:bg-blue-700"
                >
                    New Project
                </button>
            </div>

            {projects?.length === 0 ? (
                <p className="text-gray-400">No projects yet.</p>
            ) : (
                <div className="grid gap-3">
                    {projects?.map(project => (
                        <div key={project.id} className="flex items-center justify-between p-4 border
  border-gray-200 rounded-lg hover:bg-gray-50">
                            <Link
                                to={`/workspaces/${workspaceId}/projects/${project.id}/issues`}
                                className="font-medium text-gray-900 flex-1"
                            >
                                {project.name}
                            </Link>
                            <div className="flex gap-2 ml-4">
                                <button
                                    onClick={() => setEditProject(project)}
                                    className="text-sm text-gray-500 hover:text-gray-800"
                                >
                                    Edit
                                </button>
                                <button
                                    onClick={() => handleDelete(project.id)}
                                    disabled={deletingId === project.id}
                                    className="text-sm text-red-400 hover:text-red-600 disabled:opacity-50"
                                >
                                    Delete
                                </button>
                            </div>
                        </div>
                    ))}
                </div>
            )}

            {showCreate && (
                <Modal title="New Project" onClose={() => setShowCreate(false)}>
                    <ProjectForm workspaceId={workspaceId!} onClose={() => setShowCreate(false)} />
                </Modal>
            )}

            {editProject && (
                <Modal title="Edit Project" onClose={() => setEditProject(null)}>
                    <ProjectForm
                        workspaceId={workspaceId!}
                        project={editProject}
                        onClose={() => setEditProject(null)}
                    />
                </Modal>
            )}
        </div>
    )
}