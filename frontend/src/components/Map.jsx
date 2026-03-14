import { useMemo, useEffect } from 'react'
import { MapContainer, TileLayer, Marker, Tooltip, useMap, useMapEvents } from 'react-leaflet'
import { cellToBoundary, cellToLatLng } from 'h3-js'
import L from 'leaflet'
import ZoneHex from './ZoneHex'
import 'leaflet/dist/leaflet.css'

const BANGKOK_CENTER = [13.7563, 100.5018]
const MONSTER_SIZE = 28
const CHAR_SIZE = 40

function makeIcon(path, size) {
  return L.icon({
    iconUrl: path,
    iconSize: [size, size],
    iconAnchor: [size / 2, size / 2],
  })
}

const monsterIcons = {}
function getMonsterIcon(iconFile) {
  if (!monsterIcons[iconFile]) {
    monsterIcons[iconFile] = makeIcon(`/markers/${iconFile}`, MONSTER_SIZE)
  }
  return monsterIcons[iconFile]
}

function makeCharIcon(path) {
  return L.divIcon({
    html: `<div style="width:${CHAR_SIZE}px;height:${CHAR_SIZE}px;border:3px solid #facc15;border-radius:50%;background:#000;display:flex;align-items:center;justify-content:center;box-shadow:0 0 8px rgba(250,204,21,0.6)"><img src="${path}" style="width:${CHAR_SIZE - 10}px;height:${CHAR_SIZE - 10}px;border-radius:50%" /></div>`,
    iconSize: [CHAR_SIZE, CHAR_SIZE],
    iconAnchor: [CHAR_SIZE / 2, CHAR_SIZE / 2],
    className: '',
  })
}

const charIcons = [
  makeCharIcon('/markers/char-1.png'),
  makeCharIcon('/markers/char-2.png'),
]

function FitZone({ boundary }) {
  const map = useMap()

  useEffect(() => {
    if (!boundary || boundary.length === 0) return
    const bounds = L.latLngBounds(boundary)
    map.fitBounds(bounds, { padding: [20, 20] })
  }, [boundary, map])

  return null
}

function MapClickHandler({ picking, onMapClick }) {
  const map = useMap()

  useEffect(() => {
    map.getContainer().style.cursor = picking ? 'crosshair' : ''
  }, [picking, map])

  useMapEvents({
    click(e) {
      if (picking && onMapClick) {
        onMapClick(e.latlng.lat, e.latlng.lng)
      }
    },
  })

  return null
}

function Map({ zoneId, monsters = [], characters = [], picking, onMapClick }) {
  const boundary = useMemo(() => {
    if (!zoneId) return null
    return cellToBoundary(zoneId)
  }, [zoneId])

  return (
    <MapContainer
      center={BANGKOK_CENTER}
      zoom={11}
      className="w-full h-full"
      zoomControl={false}
    >
      <TileLayer
        attribution='&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>'
        url="https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png"
      />

      <MapClickHandler picking={picking} onMapClick={onMapClick} />

      {boundary && (
        <>
          <ZoneHex boundary={boundary} color="#4ade80" label={zoneId} />
          <FitZone boundary={boundary} />
        </>
      )}

      {monsters.filter(m => m.is_alive).map(m => {
        const [lat, lng] = cellToLatLng(m.h3_index)
        return (
          <Marker key={m.id} position={[lat, lng]} icon={getMonsterIcon(m.icon)}>
            <Tooltip>
              {m.type} — HP: {m.current_hp}/{m.max_hp}
            </Tooltip>
          </Marker>
        )
      })}

      {characters.map((c, i) => {
        if (c.hp <= 0) return null
        const [lat, lng] = cellToLatLng(c.h3_index)
        return (
          <Marker key={c.id} position={[lat, lng]} icon={charIcons[i % 2]}>
            <Tooltip>
              {c.name} — HP: {c.hp}/{c.max_hp}
            </Tooltip>
          </Marker>
        )
      })}
    </MapContainer>
  )
}

export default Map
