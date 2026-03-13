import { MapContainer, TileLayer } from 'react-leaflet'
import 'leaflet/dist/leaflet.css'

const BANGKOK_CENTER = [13.7563, 100.5018]

function Map() {
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

      {/* TODO: Render H3 res-6 zone hexagons as Leaflet Polygons */}
      {/* TODO: ZoneHex component with click-to-deploy */}
    </MapContainer>
  )
}

export default Map
