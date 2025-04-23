package maps

import "maps"

/**
 * MergeMaps merges two maps of the same type into a new map.
 * If m1 and m2 contain the same key, the value will be get from m2.
 *
 * Parameters:
 * - m1: the first map to merge
 * - m2: the second map to merge
 *
 * Returns:
 * A new map containing all key-value pairs from both input maps.
 */
func MergeMaps[T1 comparable, T2 any](m1 map[T1]T2, m2 map[T1]T2) map[T1]T2 {
	if m1 == nil && m2 == nil {
		return nil
	}
	res := make(map[T1]T2, len(m1)+len(m2))
	maps.Copy(res, m1)
	maps.Copy(res, m2)
	return res
}
