# entityInstances[].__tile

jq '[.defs.entities[] | {id:.identifier, tile:.tileRect}] | map(select(.id | match("(Waypoint|Map)")))' assets/maps/maps.ldtk


