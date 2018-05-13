package othello

type movements struct {
	blackCached bool
	whiteCached bool
	black []tuple
	white []tuple
}

func newMovements() *movements {
	return &movements {}
}

type MapMovementsCache map[[64]byte]*movements

func NewMapMovementsCache() MapMovementsCache {
	return make(MapMovementsCache)
}

func (c MapMovementsCache) Movements(id [64]byte, player int8) ([]tuple, bool) {

	if storedMovements, found := c[id]; found {
		if player==WHITE && storedMovements.whiteCached {
			return storedMovements.white, true
		}
		if player==WHITE && storedMovements.blackCached {
			return storedMovements.black, true
		}
	}
	return nil, false
}

func (c MapMovementsCache) StoreMovements(player int8, id [64]byte, validMovs []tuple) {
	var storedMovements *movements
	var found bool

	if storedMovements, found = c[id]; !found {
		storedMovements = newMovements()
	}
	if player==WHITE {
		storedMovements.whiteCached = true
		storedMovements.white = validMovs
	} else{
		storedMovements.blackCached = true
		storedMovements.black = validMovs
	}
	c[id] = storedMovements
}



