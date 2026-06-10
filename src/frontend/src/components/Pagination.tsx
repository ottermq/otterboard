interface Props {
    page: number
    limit: number
    total: number
    onPage: (page: number) => void
}

export default function Pagination({ page, limit, total, onPage }: Props) {
    const totalPages = Math.ceil(total / limit)
    if (totalPages <= 1) return null

    return (
        <div className="flex items-center justify-between mt-4 text-sm text-gray-600">
            <span>
                {(page - 1) * limit + 1}–{Math.min(page * limit, total)} of {total}
            </span>
            <div className="flex gap-2">
                <button
                    onClick={() => onPage(page - 1)}
                    disabled={page <= 1}
                    className="px-3 py-1.5 border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                    Previous
                </button>
                <button
                    onClick={() => onPage(page + 1)}
                    disabled={page >= totalPages}
                    className="px-3 py-1.5 border border-gray-200 rounded hover:bg-gray-50 disabled:opacity-40 disabled:cursor-not-allowed"
                >
                    Next
                </button>
            </div>
        </div>
    )
}