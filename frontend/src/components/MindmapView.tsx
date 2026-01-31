import { useState, useRef, useCallback, useMemo } from 'react'
import type { Mindmap, NoeudMindmap, LienMindmap } from '../services/api'

interface MindmapViewProps {
  mindmap: Mindmap
}

// Couleurs par type de noeud
const COULEURS_NOEUDS: Record<NoeudMindmap['type'], { bg: string; border: string; text: string }> = {
  central: { bg: '#E85D4C', border: '#C54A3C', text: '#FFFFFF' },
  branche: { bg: '#1A4D4D', border: '#153D3D', text: '#FFFFFF' },
  feuille: { bg: '#F5C542', border: '#D4A835', text: '#1A1A1A' },
}

// Dimensions des noeuds
const DIMENSIONS_NOEUDS: Record<NoeudMindmap['type'], { width: number; height: number; rx: number }> = {
  central: { width: 160, height: 60, rx: 30 },
  branche: { width: 140, height: 44, rx: 10 },
  feuille: { width: 120, height: 36, rx: 8 },
}

// Calculer le viewBox optimal pour contenir tous les noeuds
function calculerViewBoxInitial(noeuds: NoeudMindmap[]) {
  if (noeuds.length === 0) {
    return { x: 0, y: 0, width: 800, height: 600 }
  }

  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity

  noeuds.forEach(n => {
    const dims = DIMENSIONS_NOEUDS[n.type]
    minX = Math.min(minX, n.position.x - dims.width / 2)
    minY = Math.min(minY, n.position.y - dims.height / 2)
    maxX = Math.max(maxX, n.position.x + dims.width / 2)
    maxY = Math.max(maxY, n.position.y + dims.height / 2)
  })

  const padding = 60
  return {
    x: minX - padding,
    y: minY - padding,
    width: maxX - minX + padding * 2,
    height: maxY - minY + padding * 2,
  }
}

function NoeudSVG({ noeud }: { noeud: NoeudMindmap }) {
  const couleurs = COULEURS_NOEUDS[noeud.type]
  const dims = DIMENSIONS_NOEUDS[noeud.type]
  const fontSize = noeud.type === 'central' ? 14 : noeud.type === 'branche' ? 12 : 11
  const fontWeight = noeud.type === 'central' ? 700 : noeud.type === 'branche' ? 600 : 500

  // Tronquer le texte si trop long
  const maxChars = noeud.type === 'central' ? 20 : noeud.type === 'branche' ? 18 : 15
  const label = noeud.label.length > maxChars
    ? noeud.label.substring(0, maxChars - 1) + '...'
    : noeud.label

  return (
    <g transform={`translate(${noeud.position.x - dims.width / 2}, ${noeud.position.y - dims.height / 2})`}>
      <rect
        width={dims.width}
        height={dims.height}
        rx={dims.rx}
        fill={couleurs.bg}
        stroke={couleurs.border}
        strokeWidth={2}
        style={{ filter: 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))' }}
      />
      <text
        x={dims.width / 2}
        y={dims.height / 2}
        textAnchor="middle"
        dominantBaseline="middle"
        fill={couleurs.text}
        fontSize={fontSize}
        fontWeight={fontWeight}
        fontFamily="'DM Sans', sans-serif"
      >
        {label}
      </text>
    </g>
  )
}

function LienSVG({ lien, noeuds }: { lien: LienMindmap; noeuds: Map<string, NoeudMindmap> }) {
  const source = noeuds.get(lien.source)
  const target = noeuds.get(lien.target)

  if (!source || !target) return null

  // Calculer le point de départ et d'arrivée aux bords des noeuds
  const dx = target.position.x - source.position.x
  const dy = target.position.y - source.position.y
  const angle = Math.atan2(dy, dx)

  const sourceDims = DIMENSIONS_NOEUDS[source.type]
  const targetDims = DIMENSIONS_NOEUDS[target.type]

  // Points de connexion sur les bords des noeuds
  const sourceX = source.position.x + Math.cos(angle) * (sourceDims.width / 2)
  const sourceY = source.position.y + Math.sin(angle) * (sourceDims.height / 2)
  const targetX = target.position.x - Math.cos(angle) * (targetDims.width / 2)
  const targetY = target.position.y - Math.sin(angle) * (targetDims.height / 2)

  // Courbe de Bezier pour un tracé plus naturel
  const midX = (sourceX + targetX) / 2
  const midY = (sourceY + targetY) / 2
  const ctrlOffset = Math.abs(dx) > Math.abs(dy) ? dy * 0.3 : dx * 0.3

  const path = `M ${sourceX} ${sourceY} Q ${midX + ctrlOffset} ${midY - ctrlOffset} ${targetX} ${targetY}`

  return (
    <path
      d={path}
      fill="none"
      stroke="#D1D5DB"
      strokeWidth={2}
      strokeLinecap="round"
    />
  )
}

export default function MindmapView({ mindmap }: MindmapViewProps) {
  const containerRef = useRef<HTMLDivElement>(null)

  // Calculer le viewBox initial avec useMemo
  const viewBoxInitial = useMemo(() => calculerViewBoxInitial(mindmap.noeuds), [mindmap.noeuds])

  // Utiliser le pattern derived state pour reset quand la mindmap change
  const [prevMindmapId, setPrevMindmapId] = useState(mindmap.id)
  const [viewBox, setViewBox] = useState(viewBoxInitial)
  const [zoom, setZoom] = useState(1)
  const [isPanning, setIsPanning] = useState(false)
  const [panStart, setPanStart] = useState({ x: 0, y: 0 })

  // Reset quand la mindmap change (derived state pattern)
  if (mindmap.id !== prevMindmapId) {
    setPrevMindmapId(mindmap.id)
    setViewBox(viewBoxInitial)
    setZoom(1)
  }

  // Map des noeuds pour acces rapide
  const noeudsMap = useMemo(() => {
    const map = new Map<string, NoeudMindmap>()
    mindmap.noeuds.forEach(n => map.set(n.id, n))
    return map
  }, [mindmap.noeuds])

  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault()
    const delta = e.deltaY > 0 ? 1.1 : 0.9
    const newZoom = Math.min(Math.max(zoom * delta, 0.5), 3)

    // Calculer le nouveau viewBox centré sur la position de la souris
    const rect = containerRef.current?.getBoundingClientRect()
    if (rect) {
      const mouseX = e.clientX - rect.left
      const mouseY = e.clientY - rect.top
      const viewX = viewBox.x + (mouseX / rect.width) * viewBox.width
      const viewY = viewBox.y + (mouseY / rect.height) * viewBox.height

      const newWidth = viewBox.width * (newZoom / zoom)
      const newHeight = viewBox.height * (newZoom / zoom)

      setViewBox({
        x: viewX - (mouseX / rect.width) * newWidth,
        y: viewY - (mouseY / rect.height) * newHeight,
        width: newWidth,
        height: newHeight,
      })
    }
    setZoom(newZoom)
  }, [zoom, viewBox])

  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button === 0) {
      setIsPanning(true)
      setPanStart({ x: e.clientX, y: e.clientY })
    }
  }, [])

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    if (!isPanning) return

    const rect = containerRef.current?.getBoundingClientRect()
    if (rect) {
      const dx = (e.clientX - panStart.x) * (viewBox.width / rect.width)
      const dy = (e.clientY - panStart.y) * (viewBox.height / rect.height)

      setViewBox(prev => ({
        ...prev,
        x: prev.x - dx,
        y: prev.y - dy,
      }))
      setPanStart({ x: e.clientX, y: e.clientY })
    }
  }, [isPanning, panStart, viewBox])

  const handleMouseUp = useCallback(() => {
    setIsPanning(false)
  }, [])

  const handleReset = useCallback(() => {
    setViewBox(calculerViewBoxInitial(mindmap.noeuds))
    setZoom(1)
  }, [mindmap.noeuds])

  return (
    <div className="relative w-full h-full min-h-[500px]">
      {/* Controles de zoom */}
      <div className="absolute top-4 right-4 z-10 flex flex-col gap-2">
        <button
          onClick={() => {
            const newZoom = Math.min(zoom * 1.2, 3)
            setViewBox(prev => ({
              ...prev,
              x: prev.x + (prev.width - prev.width * (zoom / newZoom)) / 2,
              y: prev.y + (prev.height - prev.height * (zoom / newZoom)) / 2,
              width: prev.width * (zoom / newZoom),
              height: prev.height * (zoom / newZoom),
            }))
            setZoom(newZoom)
          }}
          className="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center text-ink hover:bg-cream transition-colors"
          title="Zoom avant"
        >
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="10" y1="5" x2="10" y2="15" />
            <line x1="5" y1="10" x2="15" y2="10" />
          </svg>
        </button>
        <button
          onClick={() => {
            const newZoom = Math.max(zoom * 0.8, 0.5)
            setViewBox(prev => ({
              ...prev,
              x: prev.x + (prev.width - prev.width * (zoom / newZoom)) / 2,
              y: prev.y + (prev.height - prev.height * (zoom / newZoom)) / 2,
              width: prev.width * (zoom / newZoom),
              height: prev.height * (zoom / newZoom),
            }))
            setZoom(newZoom)
          }}
          className="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center text-ink hover:bg-cream transition-colors"
          title="Zoom arriere"
        >
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
            <line x1="5" y1="10" x2="15" y2="10" />
          </svg>
        </button>
        <button
          onClick={handleReset}
          className="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center text-ink hover:bg-cream transition-colors"
          title="Recentrer"
        >
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="3" width="14" height="14" rx="2" />
            <circle cx="10" cy="10" r="2" fill="currentColor" />
          </svg>
        </button>
      </div>

      {/* Legende */}
      <div className="absolute bottom-4 left-4 z-10 bg-white/90 backdrop-blur-sm rounded-lg p-3 shadow-md">
        <div className="text-xs font-semibold text-ink mb-2">Legende</div>
        <div className="flex flex-col gap-1.5">
          <div className="flex items-center gap-2">
            <div className="w-4 h-4 rounded-full" style={{ backgroundColor: COULEURS_NOEUDS.central.bg }} />
            <span className="text-xs text-ink-light">Theme principal</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-4 h-3 rounded" style={{ backgroundColor: COULEURS_NOEUDS.branche.bg }} />
            <span className="text-xs text-ink-light">Sous-theme</span>
          </div>
          <div className="flex items-center gap-2">
            <div className="w-3 h-2.5 rounded" style={{ backgroundColor: COULEURS_NOEUDS.feuille.bg }} />
            <span className="text-xs text-ink-light">Detail</span>
          </div>
        </div>
      </div>

      {/* SVG Mindmap */}
      <div
        ref={containerRef}
        className="w-full h-full bg-cream rounded-lg cursor-grab active:cursor-grabbing overflow-hidden"
        onWheel={handleWheel}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
      >
        <svg
          width="100%"
          height="100%"
          viewBox={`${viewBox.x} ${viewBox.y} ${viewBox.width} ${viewBox.height}`}
          preserveAspectRatio="xMidYMid meet"
        >
          {/* Liens d'abord pour qu'ils soient sous les noeuds */}
          <g className="liens">
            {mindmap.liens.map(lien => (
              <LienSVG key={lien.id} lien={lien} noeuds={noeudsMap} />
            ))}
          </g>

          {/* Noeuds */}
          <g className="noeuds">
            {mindmap.noeuds.map(noeud => (
              <NoeudSVG key={noeud.id} noeud={noeud} />
            ))}
          </g>
        </svg>
      </div>

      {/* Instructions */}
      <div className="absolute bottom-4 right-4 text-xs text-ink-muted">
        Molette: zoom | Clic + glisser: deplacer
      </div>
    </div>
  )
}
