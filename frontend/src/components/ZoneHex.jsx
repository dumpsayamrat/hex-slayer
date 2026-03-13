import { Polygon, Tooltip } from 'react-leaflet'

// TODO: Render a single H3 res-6 zone as a Leaflet polygon
// Props: h3Index, monsterCount, hasCharacter, inCombat, onClick

function ZoneHex({ boundary, label, color = 'gray', onClick }) {
  return (
    <Polygon
      positions={boundary}
      pathOptions={{
        color,
        weight: 2,
        fillOpacity: 0.3,
      }}
      eventHandlers={{ click: onClick }}
    >
      {label && <Tooltip>{label}</Tooltip>}
    </Polygon>
  )
}

export default ZoneHex
