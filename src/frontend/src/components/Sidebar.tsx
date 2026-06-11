import { useQuery } from '@tanstack/react-query';
import { useState } from 'react';
import { NavLink, useNavigate, useParams } from 'react-router-dom';
import { listWorkspaces } from '../api/workspaces';
import { useProjects } from '../hooks/useProjecs';
import Modal from './Modal';
import ProjectForm from './ProjectForm';

export default function Sidebar() {
    const { workspaceId } = useParams<{ workspaceId: string }>()
    const navigate = useNavigate()

    const [switcherOpen, setSwitherOpen] = useState(false)
    const [showNewProject, setShowNewProject] = useState(false)

    const { data: workspaces } = useQuery({ queryKey: ['workspaces'], queryFn: listWorkspaces })
    const { data: projects } = useProjects(workspaceId!)

    const currentWorkspace = workspaces?.find(w => w.id === workspaceId)

    function switchWorkspace(id: string) {
        setSwitherOpen(false)
        navigate(`/workspaces/${id}/issues`)
    }

    const navClass = ({ isActive }: { isActive: boolean }) =>
        `flex items-center gap-3 px-3 py-2 rounded-lg text-sm transition-colors ${isActive
            ? 'bg-gray-200 text-gray-900 font-medium'
            : 'text-gray-600 hover:bg-gray-100 hover:text-gray-900'
        }`

    return (
        <aside className="w-60 h-full bg-gray-100 border-r border-gray-200 flex flex-col overflow-y-auto shrink-0">

            {/* Logo */}
            <div className="px-5 py-4">
                <span className="text-lg font-bold text-gray-900">OtterBoard</span>
            </div>

            <hr className="boarder-gray-200 mx-3" />

            {/* Workspaces */}
            <div className="px-3 pt-4 pb-2">
                <div className="flex items-center justify-between px-1 mb-2">
                    <span className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
                        Workspaces
                    </span>
                    <button className="w-5 h-5 rounded-full bg-gray-200 hover:bg-gray-300 text-gray-500 text-xs flex items-center justify-center">
                        +
                    </button>
                </div>

                <div className="relative">
                    <button
                        onClick={() => setSwitherOpen(p => !p)}
                        className="w-full flex items-center justify-between px-3 py-2 rounded-lg text-sm text-gray-700 hover:bg-gray-200"
                    >
                        <span className='truncate font-medium'>{currentWorkspace?.name ?? '...'}</span>
                        <span className='text-gray-400 text-xs ml-1'>▾</span>
                    </button>

                    {switcherOpen && (
                        <div className='absolute left-0  top-full mt-1 w-full bg-white rounded-lg shadow-lg border border-gray-200 z-10 py-1'>
                            {workspaces?.map(w => (
                                <button
                                    key={w.id}
                                    onClick={() => switchWorkspace(w.id)}
                                    className={`w-full text-left px-3 py-2 text-sm hover:bg-gray-50 ${w.id === workspaceId ? 'font-medium text-blue-600' : 'text-gray-700'
                                        }`}
                                >
                                    {w.name}
                                </button>
                            ))}
                        </div>
                    )}
                </div>
            </div>

            <hr className='border-gray-200 mx-3' />

            {/* Nav */}
            <nav className='px-3 py-3 flex flex-col gap-1'>
                <NavLink to={`/workspaces/${workspaceId}/issues?assignee=me`} className={navClass}>
                    <span>🏠</span><span>My Issues</span>
                </NavLink>
                <NavLink to={`/workspaces/${workspaceId}/issues`} end className={navClass}>
                    <span>✅</span><span>All Issues</span>
                </NavLink>
                <NavLink to={`/workspaces/${workspaceId}/members`} className={navClass}>
                    <span>👥</span><span>Members</span>
                </NavLink>
                <NavLink to={`/workspaces/${workspaceId}/settings`} className={navClass}>
                    <span>⚙️</span><span>Settings</span>
                </NavLink>
            </nav>

            <hr className='border-gray-200 mx-3' />

            {/* Projects */}
            <div className='px-3 py-2 flex flex-col flex-1'>
                <div className='flex items-center justify-between px-1 mb-2'>
                    <span className="text-xs font-semibold text-gray-400 uppercase tracking-wider">
                        Projects
                    </span>
                    <button
                        onClick={() => setShowNewProject(true)}
                        className="w-5 h-5 rounded-full bg-gray-200 hover:bg-gray-300 text-gray-500 text-xs flex items-center justify-center"
                    >
                        +
                    </button>
                </div>

                <div className='flex flex-col gap-1'>
                    {projects?.map(project => (
                        <NavLink
                            key={project.id}
                            to={`/workspaces/${workspaceId}/projects/${project.id}/issues`}
                            className={navClass}
                        >
                            <span className='w-6 h-6 rounded bg-indigo-100 text-indigo-700 text-xs font-bold flex items-center justify-center shrink-0'>
                                {project.name[0].toUpperCase()}
                            </span>
                            <span className='truncate'>{project.name}</span>
                        </NavLink>
                    ))}
                </div>
            </div>

            {showNewProject && (
                <Modal title='New Project' onClose={() => setShowNewProject(false)}>
                    <ProjectForm workspaceId={workspaceId!} onClose={() => setShowNewProject(false)} />
                </Modal>
            )}
        </aside>
    )
}