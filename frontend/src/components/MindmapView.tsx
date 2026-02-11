import { useState, useRef, useCallback, useMemo, useEffect } from 'react'
import type { Mindmap, NoeudMindmap, LienMindmap, Position } from '../services/api'

interface MindmapViewProps {
  mindmap: Mindmap
  onNoeudClick?: (noeud: NoeudMindmap) => void
}

interface ViewBox {
  x: number
  y: number
  width: number
  height: number
}

interface DragInfo {
  noeudId: string
  startSvgX: number
  startSvgY: number
  startNodeX: number
  startNodeY: number
  hasMoved: boolean
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

const DRAG_THRESHOLD = 5

const MAX_CHARS: Record<NoeudMindmap['type'], number> = {
  central: 20,
  branche: 18,
  feuille: 15,
}

function estTronque(noeud: NoeudMindmap): boolean {
  return noeud.label.length > MAX_CHARS[noeud.type]
}

function tronquerLabel(noeud: NoeudMindmap): string {
  const max = MAX_CHARS[noeud.type]
  return noeud.label.length > max
    ? noeud.label.substring(0, max - 1) + '...'
    : noeud.label
}

// Convertir coordonnees client en coordonnees SVG
function clientToSvg(clientX: number, clientY: number, rect: DOMRect, vb: ViewBox): Position {
  return {
    x: vb.x + ((clientX - rect.left) / rect.width) * vb.width,
    y: vb.y + ((clientY - rect.top) / rect.height) * vb.height,
  }
}

// Calculer le viewBox optimal pour contenir les noeuds visibles
function calculerViewBox(noeuds: NoeudMindmap[], positions: Record<string, Position>): ViewBox {
  if (noeuds.length === 0) {
    return { x: 0, y: 0, width: 800, height: 600 }
  }

  let minX = Infinity, minY = Infinity, maxX = -Infinity, maxY = -Infinity

  noeuds.forEach(n => {
    const pos = positions[n.id] || n.position
    const dims = DIMENSIONS_NOEUDS[n.type]
    minX = Math.min(minX, pos.x - dims.width / 2)
    minY = Math.min(minY, pos.y - dims.height / 2)
    maxX = Math.max(maxX, pos.x + dims.width / 2)
    maxY = Math.max(maxY, pos.y + dims.height / 2)
  })

  const padding = 60
  return {
    x: minX - padding,
    y: minY - padding,
    width: maxX - minX + padding * 2,
    height: maxY - minY + padding * 2,
  }
}

interface NoeudSVGProps {
  noeud: NoeudMindmap
  position: Position
  hasEnfants: boolean
  isReplie: boolean
  isDragging: boolean
  onPointerDown: (e: React.MouseEvent | React.TouchEvent, noeudId: string) => void
  onToggle: (noeudId: string) => void
  onHover: (noeudId: string | null) => void
}

function NoeudSVG({ noeud, position, hasEnfants, isReplie, isDragging, onPointerDown, onToggle, onHover }: NoeudSVGProps) {
  const [isHovered, setIsHovered] = useState(false)
  const couleurs = COULEURS_NOEUDS[noeud.type]
  const dims = DIMENSIONS_NOEUDS[noeud.type]
  const fontSize = noeud.type === 'central' ? 14 : noeud.type === 'branche' ? 12 : 11
  const fontWeight = noeud.type === 'central' ? 700 : noeud.type === 'branche' ? 600 : 500
  const isClickable = !!noeud.conceptId
  const label = tronquerLabel(noeud)
  const tronque = estTronque(noeud)

  const handleMouseDown = (e: React.MouseEvent) => {
    e.stopPropagation()
    onPointerDown(e, noeud.id)
  }

  const handleTouchStart = (e: React.TouchEvent) => {
    e.stopPropagation()
    onPointerDown(e, noeud.id)
    // Montrer le tooltip au toucher pour mobile
    if (tronque) onHover(noeud.id)
  }

  const handleMouseEnter = () => {
    setIsHovered(true)
    if (tronque) onHover(noeud.id)
  }

  const handleMouseLeave = () => {
    setIsHovered(false)
    onHover(null)
  }

  // Toggle collapse - stop all propagation
  const handleToggleMouseDown = (e: React.MouseEvent) => { e.stopPropagation() }
  const handleToggleTouchStart = (e: React.TouchEvent) => { e.stopPropagation() }
  const handleToggleClick = (e: React.MouseEvent) => {
    e.stopPropagation()
    onToggle(noeud.id)
  }

  return (
    <g
      transform={`translate(${position.x - dims.width / 2}, ${position.y - dims.height / 2})`}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      onMouseDown={handleMouseDown}
      onTouchStart={handleTouchStart}
      style={{ cursor: isDragging ? 'grabbing' : 'grab' }}
    >
      <rect
        width={dims.width}
        height={dims.height}
        rx={dims.rx}
        fill={couleurs.bg}
        stroke={
          isDragging ? '#F5C542'
            : isReplie ? '#F5C542'
              : isClickable && isHovered ? '#F5C542'
                : couleurs.border
        }
        strokeWidth={isDragging ? 3 : isClickable && isHovered ? 3 : 2}
        style={{
          filter: isDragging
            ? 'drop-shadow(0 4px 12px rgba(0,0,0,0.3))'
            : isClickable && isHovered
              ? 'drop-shadow(0 2px 8px rgba(245,197,66,0.4))'
              : 'drop-shadow(0 2px 4px rgba(0,0,0,0.1))',
          opacity: isReplie ? 0.85 : 1,
        }}
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
        style={{
          textDecoration: isClickable ? 'underline' : 'none',
          pointerEvents: 'none',
          userSelect: 'none',
        }}
      >
        {label}
      </text>

      {/* Indicateur de lien vers concept */}
      {isClickable && !hasEnfants && (
        <g transform={`translate(${dims.width - 16}, 4)`}>
          <circle cx="6" cy="6" r="6" fill="white" opacity="0.85" />
          <path
            d="M4 8L8 4M8 4H5.5M8 4V6.5"
            stroke={couleurs.bg}
            strokeWidth="1.5"
            strokeLinecap="round"
            strokeLinejoin="round"
            fill="none"
          />
        </g>
      )}

      {/* Bouton replier/deplier */}
      {hasEnfants && (
        <g
          transform={`translate(${dims.width - 2}, ${dims.height / 2})`}
          onClick={handleToggleClick}
          onMouseDown={handleToggleMouseDown}
          onTouchStart={handleToggleTouchStart}
          style={{ cursor: 'pointer' }}
        >
          <rect x={-12} y={-12} width={24} height={24} fill="transparent" />
          <circle
            r={10}
            fill="white"
            stroke={isReplie ? '#F5C542' : couleurs.border}
            strokeWidth={1.5}
            style={{ filter: 'drop-shadow(0 1px 2px rgba(0,0,0,0.15))' }}
          />
          {isReplie ? (
            <>
              <line x1={-4} y1={0} x2={4} y2={0} stroke={couleurs.bg} strokeWidth={2} strokeLinecap="round" />
              <line x1={0} y1={-4} x2={0} y2={4} stroke={couleurs.bg} strokeWidth={2} strokeLinecap="round" />
            </>
          ) : (
            <line x1={-4} y1={0} x2={4} y2={0} stroke={couleurs.bg} strokeWidth={2} strokeLinecap="round" />
          )}
        </g>
      )}
    </g>
  )
}

function LienSVG({ lien, noeudsMap }: { lien: LienMindmap; noeudsMap: Map<string, { type: NoeudMindmap['type']; position: Position }> }) {
  const source = noeudsMap.get(lien.source)
  const target = noeudsMap.get(lien.target)

  if (!source || !target) return null

  const dx = target.position.x - source.position.x
  const dy = target.position.y - source.position.y
  const angle = Math.atan2(dy, dx)

  const sourceDims = DIMENSIONS_NOEUDS[source.type]
  const targetDims = DIMENSIONS_NOEUDS[target.type]

  const sourceX = source.position.x + Math.cos(angle) * (sourceDims.width / 2)
  const sourceY = source.position.y + Math.sin(angle) * (sourceDims.height / 2)
  const targetX = target.position.x - Math.cos(angle) * (targetDims.width / 2)
  const targetY = target.position.y - Math.sin(angle) * (targetDims.height / 2)

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

export default function MindmapView({ mindmap, onNoeudClick }: MindmapViewProps) {
  const containerRef = useRef<HTMLDivElement>(null)
  const tooltipTimeoutRef = useRef<number | null>(null)
  const pinchRef = useRef<{ startDist: number; startViewBox: ViewBox; centerX: number; centerY: number } | null>(null)

  // Positions mutables des noeuds (pour drag)
  const [positions, setPositions] = useState<Record<string, Position>>(() => {
    const pos: Record<string, Position> = {}
    mindmap.noeuds.forEach(n => { pos[n.id] = { ...n.position } })
    return pos
  })

  // Noeuds replies (collapsed)
  const [repliedIds, setRepliedIds] = useState<Set<string>>(new Set())

  // Drag
  const [dragInfo, setDragInfo] = useState<DragInfo | null>(null)

  // Tooltip
  const [tooltipNoeudId, setTooltipNoeudId] = useState<string | null>(null)

  // ViewBox, zoom, pan
  const viewBoxInitial = useMemo(
    () => calculerViewBox(mindmap.noeuds, positions),
    // eslint-disable-next-line react-hooks/exhaustive-deps
    [mindmap.id],
  )
  const [prevMindmapId, setPrevMindmapId] = useState(mindmap.id)
  const [viewBox, setViewBox] = useState(viewBoxInitial)
  const [zoom, setZoom] = useState(1)
  const [isPanning, setIsPanning] = useState(false)
  const [panStart, setPanStart] = useState({ x: 0, y: 0 })

  // Cleanup timeout on unmount
  useEffect(() => {
    return () => {
      if (tooltipTimeoutRef.current) clearTimeout(tooltipTimeoutRef.current)
    }
  }, [])

  // Reset quand la mindmap change
  if (mindmap.id !== prevMindmapId) {
    setPrevMindmapId(mindmap.id)
    const pos: Record<string, Position> = {}
    mindmap.noeuds.forEach(n => { pos[n.id] = { ...n.position } })
    setPositions(pos)
    setRepliedIds(new Set())
    setDragInfo(null)
    setTooltipNoeudId(null)
    const newVb = calculerViewBox(mindmap.noeuds, pos)
    setViewBox(newVb)
    setZoom(1)
  }

  // Arbre parent → enfants a partir des liens
  const { enfantsMap, noeudsAvecEnfants } = useMemo(() => {
    const map = new Map<string, string[]>()
    mindmap.liens.forEach(l => {
      const list = map.get(l.source) || []
      list.push(l.target)
      map.set(l.source, list)
    })
    return { enfantsMap: map, noeudsAvecEnfants: new Set(map.keys()) }
  }, [mindmap.liens])

  // Calculer tous les descendants d'un noeud (iteratif)
  const getDescendants = useCallback((id: string): Set<string> => {
    const result = new Set<string>()
    const stack = [...(enfantsMap.get(id) || [])]
    while (stack.length > 0) {
      const current = stack.pop()!
      if (!result.has(current)) {
        result.add(current)
        const children = enfantsMap.get(current) || []
        stack.push(...children)
      }
    }
    return result
  }, [enfantsMap])

  // Noeuds caches (descendants de noeuds replies)
  const noeudsCache = useMemo(() => {
    const hidden = new Set<string>()
    repliedIds.forEach(id => {
      getDescendants(id).forEach(d => hidden.add(d))
    })
    return hidden
  }, [repliedIds, getDescendants])

  // Filtrer noeuds et liens visibles
  const noeudsVisibles = useMemo(
    () => mindmap.noeuds.filter(n => !noeudsCache.has(n.id)),
    [mindmap.noeuds, noeudsCache],
  )

  const liensVisibles = useMemo(
    () => mindmap.liens.filter(l => !noeudsCache.has(l.source) && !noeudsCache.has(l.target)),
    [mindmap.liens, noeudsCache],
  )

  // Map des noeuds avec positions a jour (pour LienSVG)
  const noeudsMapAvecPos = useMemo(() => {
    const map = new Map<string, { type: NoeudMindmap['type']; position: Position }>()
    noeudsVisibles.forEach(n => {
      map.set(n.id, { type: n.type, position: positions[n.id] || n.position })
    })
    return map
  }, [noeudsVisibles, positions])

  // Nombre de noeuds caches par noeud replie
  const compteurCaches = useMemo(() => {
    const counts: Record<string, number> = {}
    repliedIds.forEach(id => {
      counts[id] = getDescendants(id).size
    })
    return counts
  }, [repliedIds, getDescendants])

  // Tooltip helpers avec debounce pour eviter le flickering
  const showTooltip = useCallback((id: string | null) => {
    if (tooltipTimeoutRef.current) {
      clearTimeout(tooltipTimeoutRef.current)
      tooltipTimeoutRef.current = null
    }
    if (id) {
      setTooltipNoeudId(id)
    } else {
      tooltipTimeoutRef.current = window.setTimeout(() => {
        setTooltipNoeudId(null)
      }, 150)
    }
  }, [])

  // Toggle repliement
  const handleToggle = useCallback((noeudId: string) => {
    setRepliedIds(prev => {
      const next = new Set(prev)
      if (next.has(noeudId)) {
        next.delete(noeudId)
      } else {
        next.add(noeudId)
      }
      return next
    })
  }, [])

  // Demarrer le drag d'un noeud
  const handleNoeudPointerDown = useCallback((e: React.MouseEvent | React.TouchEvent, noeudId: string) => {
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    let clientX: number, clientY: number
    if ('touches' in e) {
      clientX = e.touches[0].clientX
      clientY = e.touches[0].clientY
    } else {
      if (e.button !== 0) return
      clientX = e.clientX
      clientY = e.clientY
    }

    const svgPos = clientToSvg(clientX, clientY, rect, viewBox)
    const nodePos = positions[noeudId]
    if (!nodePos) return

    setDragInfo({
      noeudId,
      startSvgX: svgPos.x,
      startSvgY: svgPos.y,
      startNodeX: nodePos.x,
      startNodeY: nodePos.y,
      hasMoved: false,
    })
  }, [viewBox, positions])

  // Zoom molette
  const handleWheel = useCallback((e: React.WheelEvent) => {
    e.preventDefault()
    setTooltipNoeudId(null) // masquer tooltip pendant zoom
    const delta = e.deltaY > 0 ? 1.1 : 0.9
    const newZoom = Math.min(Math.max(zoom * delta, 0.5), 3)

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

  // Mouse handlers
  const handleMouseDown = useCallback((e: React.MouseEvent) => {
    if (e.button === 0 && !dragInfo) {
      setIsPanning(true)
      setPanStart({ x: e.clientX, y: e.clientY })
      setTooltipNoeudId(null) // masquer tooltip pendant pan
    }
  }, [dragInfo])

  const handleMouseMove = useCallback((e: React.MouseEvent) => {
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    if (dragInfo) {
      const svgPos = clientToSvg(e.clientX, e.clientY, rect, viewBox)
      const dx = svgPos.x - dragInfo.startSvgX
      const dy = svgPos.y - dragInfo.startSvgY
      const dist = Math.sqrt(dx * dx + dy * dy)

      if (!dragInfo.hasMoved && dist < DRAG_THRESHOLD) return

      if (!dragInfo.hasMoved) setTooltipNoeudId(null) // masquer au debut du drag

      setDragInfo(prev => prev ? { ...prev, hasMoved: true } : null)
      setPositions(prev => ({
        ...prev,
        [dragInfo.noeudId]: {
          x: dragInfo.startNodeX + dx,
          y: dragInfo.startNodeY + dy,
        },
      }))
    } else if (isPanning) {
      const dxClient = (e.clientX - panStart.x) * (viewBox.width / rect.width)
      const dyClient = (e.clientY - panStart.y) * (viewBox.height / rect.height)

      setViewBox(prev => ({
        ...prev,
        x: prev.x - dxClient,
        y: prev.y - dyClient,
      }))
      setPanStart({ x: e.clientX, y: e.clientY })
    }
  }, [dragInfo, isPanning, panStart, viewBox])

  const handleMouseUp = useCallback(() => {
    if (dragInfo) {
      if (!dragInfo.hasMoved && onNoeudClick) {
        const noeud = mindmap.noeuds.find(n => n.id === dragInfo.noeudId)
        if (noeud?.conceptId) onNoeudClick(noeud)
      }
      setDragInfo(null)
    }
    setIsPanning(false)
  }, [dragInfo, onNoeudClick, mindmap.noeuds])

  // Touch handlers
  const handleTouchStart = useCallback((e: React.TouchEvent) => {
    if (e.touches.length === 2) {
      // Pinch-to-zoom: annuler pan/drag en cours
      setIsPanning(false)
      setDragInfo(null)
      setTooltipNoeudId(null)

      const t0 = e.touches[0], t1 = e.touches[1]
      const dist = Math.hypot(t1.clientX - t0.clientX, t1.clientY - t0.clientY)
      const centerX = (t0.clientX + t1.clientX) / 2
      const centerY = (t0.clientY + t1.clientY) / 2
      pinchRef.current = { startDist: dist, startViewBox: { ...viewBox }, centerX, centerY }
    } else if (e.touches.length === 1 && !dragInfo) {
      pinchRef.current = null
      const touch = e.touches[0]
      setIsPanning(true)
      setPanStart({ x: touch.clientX, y: touch.clientY })
    }
  }, [dragInfo, viewBox])

  const handleTouchMove = useCallback((e: React.TouchEvent) => {
    // Pinch-to-zoom
    if (e.touches.length === 2 && pinchRef.current) {
      e.preventDefault()
      const t0 = e.touches[0], t1 = e.touches[1]
      const dist = Math.hypot(t1.clientX - t0.clientX, t1.clientY - t0.clientY)
      const scale = pinchRef.current.startDist / dist

      const rect = containerRef.current?.getBoundingClientRect()
      if (rect) {
        const newWidth = pinchRef.current.startViewBox.width * scale
        const newHeight = pinchRef.current.startViewBox.height * scale

        // Limiter le zoom (0.25x a 3x)
        const origW = pinchRef.current.startViewBox.width
        if (newWidth < origW * 0.25 || newWidth > origW * 3) return

        const svgCenter = clientToSvg(
          pinchRef.current.centerX, pinchRef.current.centerY,
          rect, pinchRef.current.startViewBox,
        )
        const fracX = (pinchRef.current.centerX - rect.left) / rect.width
        const fracY = (pinchRef.current.centerY - rect.top) / rect.height

        setViewBox({
          x: svgCenter.x - fracX * newWidth,
          y: svgCenter.y - fracY * newHeight,
          width: newWidth,
          height: newHeight,
        })
      }
      return
    }

    if (e.touches.length !== 1) return
    const touch = e.touches[0]
    const rect = containerRef.current?.getBoundingClientRect()
    if (!rect) return

    if (dragInfo) {
      e.preventDefault()
      const svgPos = clientToSvg(touch.clientX, touch.clientY, rect, viewBox)
      const dx = svgPos.x - dragInfo.startSvgX
      const dy = svgPos.y - dragInfo.startSvgY
      const dist = Math.sqrt(dx * dx + dy * dy)

      if (!dragInfo.hasMoved && dist < DRAG_THRESHOLD) return

      if (!dragInfo.hasMoved) setTooltipNoeudId(null)

      setDragInfo(prev => prev ? { ...prev, hasMoved: true } : null)
      setPositions(prev => ({
        ...prev,
        [dragInfo.noeudId]: {
          x: dragInfo.startNodeX + dx,
          y: dragInfo.startNodeY + dy,
        },
      }))
    } else if (isPanning) {
      setTooltipNoeudId(null) // masquer pendant pan tactile
      const dxClient = (touch.clientX - panStart.x) * (viewBox.width / rect.width)
      const dyClient = (touch.clientY - panStart.y) * (viewBox.height / rect.height)

      setViewBox(prev => ({
        ...prev,
        x: prev.x - dxClient,
        y: prev.y - dyClient,
      }))
      setPanStart({ x: touch.clientX, y: touch.clientY })
    }
  }, [dragInfo, isPanning, panStart, viewBox])

  const handleTouchEnd = useCallback(() => {
    if (pinchRef.current) {
      pinchRef.current = null
      return
    }
    if (dragInfo) {
      if (!dragInfo.hasMoved && onNoeudClick) {
        const noeud = mindmap.noeuds.find(n => n.id === dragInfo.noeudId)
        if (noeud?.conceptId) onNoeudClick(noeud)
      }
      setDragInfo(null)
    }
    setTooltipNoeudId(null)
    setIsPanning(false)
  }, [dragInfo, onNoeudClick, mindmap.noeuds])

  const handleReset = useCallback(() => {
    const pos: Record<string, Position> = {}
    mindmap.noeuds.forEach(n => { pos[n.id] = { ...n.position } })
    setPositions(pos)
    setRepliedIds(new Set())
    setTooltipNoeudId(null)
    setViewBox(calculerViewBox(mindmap.noeuds, pos))
    setZoom(1)
  }, [mindmap.noeuds])

  // Replier/deplier tout
  const handleToggleTout = useCallback(() => {
    if (repliedIds.size > 0) {
      setRepliedIds(new Set())
    } else {
      const toReplie = new Set<string>()
      noeudsAvecEnfants.forEach(id => toReplie.add(id))
      const central = mindmap.noeuds.find(n => n.type === 'central')
      if (central) toReplie.delete(central.id)
      setRepliedIds(toReplie)
    }
  }, [repliedIds, noeudsAvecEnfants, mindmap.noeuds])

  // Calcul du tooltip HTML positionne au-dessus du noeud
  let tooltipContent: React.ReactNode = null
  if (tooltipNoeudId) {
    const noeud = mindmap.noeuds.find(n => n.id === tooltipNoeudId)
    const pos = noeud ? positions[tooltipNoeudId] : null
    const rect = containerRef.current?.getBoundingClientRect()
    if (noeud && pos && rect && rect.width > 0) {
      const pixelX = ((pos.x - viewBox.x) / viewBox.width) * rect.width
      const pixelY = ((pos.y - viewBox.y) / viewBox.height) * rect.height
      const dims = DIMENSIONS_NOEUDS[noeud.type]
      const nodeHalfH = (dims.height / 2 / viewBox.height) * rect.height
      const couleurs = COULEURS_NOEUDS[noeud.type]

      // Positionner au-dessus sauf si pas assez de place
      const above = pixelY - nodeHalfH > 70
      const clampedX = Math.max(130, Math.min(pixelX, rect.width - 130))

      tooltipContent = (
        <div
          className="absolute z-20 pointer-events-none"
          style={{
            left: clampedX,
            top: above ? pixelY - nodeHalfH - 10 : pixelY + nodeHalfH + 10,
            transform: above ? 'translate(-50%, -100%)' : 'translate(-50%, 0%)',
          }}
        >
          {/* Fleche vers le bas (tooltip au-dessus) */}
          {above && (
            <div className="flex flex-col items-center">
              <div
                className="rounded-lg shadow-lg px-4 py-2.5 max-w-[260px]"
                style={{
                  background: 'white',
                  borderLeft: `3px solid ${couleurs.bg}`,
                }}
              >
                <p className="text-sm text-ink font-medium leading-snug break-words">{noeud.label}</p>
                {noeud.conceptId && (
                  <p className="text-[11px] text-coral mt-1.5 font-medium">Cliquer pour voir le concept</p>
                )}
              </div>
              <svg width="14" height="8" viewBox="0 0 14 8" className="-mt-[1px]">
                <path d="M0 0L7 8L14 0Z" fill="white" />
                <path d="M0 0L7 7L14 0" fill="none" stroke="#e5e2de" strokeWidth="1" />
              </svg>
            </div>
          )}
          {/* Fleche vers le haut (tooltip en-dessous) */}
          {!above && (
            <div className="flex flex-col items-center">
              <svg width="14" height="8" viewBox="0 0 14 8" className="-mb-[1px]">
                <path d="M0 8L7 0L14 8Z" fill="white" />
                <path d="M0 8L7 1L14 8" fill="none" stroke="#e5e2de" strokeWidth="1" />
              </svg>
              <div
                className="rounded-lg shadow-lg px-4 py-2.5 max-w-[260px]"
                style={{
                  background: 'white',
                  borderLeft: `3px solid ${couleurs.bg}`,
                }}
              >
                <p className="text-sm text-ink font-medium leading-snug break-words">{noeud.label}</p>
                {noeud.conceptId && (
                  <p className="text-[11px] text-coral mt-1.5 font-medium">Cliquer pour voir le concept</p>
                )}
              </div>
            </div>
          )}
        </div>
      )
    }
  }

  return (
    <div className="relative w-full h-full min-h-[500px]">
      {/* Controles */}
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
          title="Reinitialiser"
        >
          <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
            <rect x="3" y="3" width="14" height="14" rx="2" />
            <circle cx="10" cy="10" r="2" fill="currentColor" />
          </svg>
        </button>
        {noeudsAvecEnfants.size > 0 && (
          <button
            onClick={handleToggleTout}
            className="w-10 h-10 bg-white rounded-lg shadow-md flex items-center justify-center text-ink hover:bg-cream transition-colors"
            title={repliedIds.size > 0 ? 'Tout deplier' : 'Tout replier'}
          >
            <svg width="20" height="20" viewBox="0 0 20 20" fill="none" stroke="currentColor" strokeWidth="2">
              {repliedIds.size > 0 ? (
                <>
                  <rect x="3" y="3" width="6" height="6" rx="1" />
                  <rect x="11" y="3" width="6" height="6" rx="1" />
                  <rect x="3" y="11" width="6" height="6" rx="1" />
                  <rect x="11" y="11" width="6" height="6" rx="1" />
                </>
              ) : (
                <rect x="4" y="4" width="12" height="12" rx="2" />
              )}
            </svg>
          </button>
        )}
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
        {noeudsCache.size > 0 && (
          <div className="mt-2 pt-2 border-t border-cream text-xs text-ink-muted">
            {noeudsCache.size} noeud{noeudsCache.size > 1 ? 's' : ''} masque{noeudsCache.size > 1 ? 's' : ''}
          </div>
        )}
      </div>

      {/* SVG Mindmap */}
      <div
        ref={containerRef}
        className="w-full h-full bg-cream rounded-lg overflow-hidden touch-none"
        style={{ cursor: dragInfo?.hasMoved ? 'grabbing' : isPanning ? 'grabbing' : 'grab' }}
        onWheel={handleWheel}
        onMouseDown={handleMouseDown}
        onMouseMove={handleMouseMove}
        onMouseUp={handleMouseUp}
        onMouseLeave={handleMouseUp}
        onTouchStart={handleTouchStart}
        onTouchMove={handleTouchMove}
        onTouchEnd={handleTouchEnd}
      >
        <svg
          width="100%"
          height="100%"
          viewBox={`${viewBox.x} ${viewBox.y} ${viewBox.width} ${viewBox.height}`}
          preserveAspectRatio="xMidYMid meet"
        >
          {/* Liens */}
          <g className="liens">
            {liensVisibles.map(lien => (
              <LienSVG key={lien.id} lien={lien} noeudsMap={noeudsMapAvecPos} />
            ))}
          </g>

          {/* Badge compteur de noeuds caches */}
          {repliedIds.size > 0 && Array.from(repliedIds).map(id => {
            const pos = positions[id]
            const count = compteurCaches[id]
            if (!pos || !count) return null
            const noeud = mindmap.noeuds.find(n => n.id === id)
            if (!noeud) return null
            const dims = DIMENSIONS_NOEUDS[noeud.type]
            return (
              <g key={`badge-${id}`} transform={`translate(${pos.x + dims.width / 2 + 8}, ${pos.y - dims.height / 2 - 4})`}>
                <rect x={-12} y={-8} width={24} height={16} rx={8} fill="#F5C542" />
                <text
                  textAnchor="middle"
                  dominantBaseline="middle"
                  fontSize={9}
                  fontWeight={700}
                  fontFamily="'DM Sans', sans-serif"
                  fill="#1A1A1A"
                  style={{ pointerEvents: 'none' }}
                >
                  +{count}
                </text>
              </g>
            )
          })}

          {/* Noeuds */}
          <g className="noeuds">
            {noeudsVisibles.map(noeud => (
              <NoeudSVG
                key={noeud.id}
                noeud={noeud}
                position={positions[noeud.id] || noeud.position}
                hasEnfants={noeudsAvecEnfants.has(noeud.id)}
                isReplie={repliedIds.has(noeud.id)}
                isDragging={dragInfo?.noeudId === noeud.id && !!dragInfo?.hasMoved}
                onPointerDown={handleNoeudPointerDown}
                onToggle={handleToggle}
                onHover={showTooltip}
              />
            ))}
          </g>
        </svg>
      </div>

      {/* Tooltip HTML au-dessus du SVG */}
      {tooltipContent}

      {/* Instructions */}
      <div className="absolute bottom-4 right-4 text-xs text-ink-muted hidden sm:block">
        Glisser: deplacer un noeud | Molette: zoom | Fond: panoramique
      </div>
    </div>
  )
}
