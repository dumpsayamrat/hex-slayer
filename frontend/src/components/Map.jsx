import { useMemo, useEffect } from 'react'
import { MapContainer, TileLayer, Marker, Tooltip, useMap, useMapEvents } from 'react-leaflet'
import { cellToBoundary, cellToLatLng } from 'h3-js'
import L from 'leaflet'
import ZoneHex from './ZoneHex'
import 'leaflet/dist/leaflet.css'

const BANGKOK_CENTER = [13.7563, 100.5018]
const MARKER_SIZE = 28

function makeIcon(path) {
  return L.icon({
    iconUrl: path,
    iconSize: [MARKER_SIZE, MARKER_SIZE],
    iconAnchor: [MARKER_SIZE / 2, MARKER_SIZE / 2],
  })
}

const monsterIcons = {}
function getMonsterIcon(iconFile) {
  if (!monsterIcons[iconFile]) {
    monsterIcons[iconFile] = makeIcon(`/markers/${iconFile}`)
  }
  return monsterIcons[iconFile]
}

const charIcons = [
  makeIcon('/markers/char-1.png'),
  makeIcon('/markers/char-2.png'),
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

function Map({ zone, picking, onMapClick }) {
  const boundary = useMemo(() => {
    if (!zone?.h3_zone) return null
    return cellToBoundary(zone.h3_zone)
  }, [zone?.h3_zone])

  const monsters = zone?.monsters || []
  const characters = zone?.characters || []

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
          <ZoneHex boundary={boundary} color="#4ade80" label={zone.h3_zone} />
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
