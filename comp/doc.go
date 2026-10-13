// Package comp is how a template of entities names one component and where its value comes
// from: the same for every entity made from it ([Const]), read from each entity's own row
// ([Load]), the tags of one family ([Tagged]), the markers of one ([Marks]), or a default of the
// store's roster left out ([Without]). Nothing here is implemented by a user; a store's binding
// spawns through a Comp's Fill, into the place it laid out.
package comp
