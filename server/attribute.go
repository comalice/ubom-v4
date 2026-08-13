package ubom

/* Attributes: planned for release MVP+1.

An attribute is a type+value pair used to give taxonomy
levels and/or part numbers canonical values. For example, we might have a taxonomy for discrete
components, where the resistors taxonomy category may have (ohm,resistance), and (string,package).

Attributes are applied in a layered fashion with naive overrides, where each lower level proceeds
to override previous levels for colliding (key,value) pairs, and PartNumber takes ultimate precidence.

*/
