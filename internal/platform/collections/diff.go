package collections

// DiffResult represents the result of comparing two collections.
type DiffResult[T any] struct {
	Added   []*T
	Updated []*T
	Deleted []*T
}

// Merged returns all items (added and updated) as a single slice of values.
func (r *DiffResult[T]) Merged() []T {
	all := make([]T, 0, len(r.Added)+len(r.Updated))
	for _, item := range r.Added {
		all = append(all, *item)
	}
	for _, item := range r.Updated {
		all = append(all, *item)
	}
	return all
}

// IdentifiableDiffer provides methods for diffing collections of items that have unique identifiers.
type IdentifiableDiffer[K comparable, T any] struct {
	// GetKey extracts the unique identifier from an item
	GetKey func(*T) K
	// AddItem is called when a new item needs to be added that doesn't exist in the existing collection.
	// it receives the existing items and the adding item and should return the added item
	// and optionally items to be updated/deleted (for complex add scenarios like replacements).
	AddItem func(existings []T, adding *T) (added *T, toUpdate *T, toDelete *T, err error)
	// UpdateItem is called when an existing item needs to be updated with new values.
	// it receives the existing item and the new item, and should return the updated item
	// and optionally items to be added/deleted (for complex update scenarios like replacements).
	UpdateItem func(existing *T, updating *T) (updated *T, toAdd *T, toDelete *T, err error)
	// DeleteItem is called when an existing item needs to be deleted that doesn't exist in the new collection.
	// it receives the new items and the deleting item and should return the deleted item
	// and optionally items to be added/updated (for complex delete scenarios like replacements).
	DeleteItem func(new []T, deleting *T) (deleted *T, toAdd *T, toUpdate *T, err error)
}

// Diff compares existing items with new items and returns what needs to be added, updated, or deleted.
// - existing: the current collection of items
// - new: the desired collection of items
func (d *IdentifiableDiffer[K, T]) Diff(existing []T, new []T) (*DiffResult[T], error) {
	var added []*T
	var updated []*T
	var deleted []*T

	// build a map of existing items by their key
	existingMap := make(map[K]*T)
	for i := range existing {
		key := d.GetKey(&existing[i])
		existingMap[key] = &existing[i]
	}

	// process new items
	for i := range new {
		newItem := &new[i]
		key := d.GetKey(newItem)

		if existingItem, ok := existingMap[key]; ok {
			// item exists - update it
			delete(existingMap, key)

			if d.UpdateItem != nil {
				updatedItem, toAdd, toDelete, err := d.UpdateItem(existingItem, newItem)
				if err != nil {
					return nil, err
				}

				if updatedItem != nil {
					updated = append(updated, updatedItem)
				}
				if toAdd != nil {
					added = append(added, toAdd)
				}
				if toDelete != nil {
					deleted = append(deleted, toDelete)
				}
			} else {
				// default: just mark as updated
				updated = append(updated, newItem)
			}
		} else {
			// item doesn't exist - add it
			if d.AddItem != nil {
				addedItem, toUpdate, toDelete, err := d.AddItem(existing, newItem)
				if err != nil {
					return nil, err
				}

				if addedItem != nil {
					added = append(added, addedItem)
				}
				if toUpdate != nil {
					updated = append(updated, toUpdate)
				}
				if toDelete != nil {
					deleted = append(deleted, toDelete)
				}
			} else {
				// default: just mark as added
				added = append(added, newItem)
			}
		}
	}

	// remaining items in existingMap are deleted
	for _, item := range existingMap {
		if d.DeleteItem != nil {
			deletedItem, toAdd, toUpdate, err := d.DeleteItem(new, item)
			if err != nil {
				return nil, err
			}

			if deletedItem != nil {
				deleted = append(deleted, deletedItem)
			}
			if toAdd != nil {
				added = append(added, toAdd)
			}
			if toUpdate != nil {
				updated = append(updated, toUpdate)
			}
		} else {
			// default: just mark as deleted
			deleted = append(deleted, item)
		}
	}

	return &DiffResult[T]{
		Added:   added,
		Updated: updated,
		Deleted: deleted,
	}, nil
}
