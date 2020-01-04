// Package mmcif takes a reader and a list of items you are interested in.
//
// It reads these lines/tables and gives them to a function which puts them
// in arrays for the pdb package.



// Package mmcif reads a file in mmcif/cif format.
// Reading mmcif files is interesting because they are so big,
// but we do not want much information from them.
// We could tokenise everything.
// If one looks at the format there are some features that make it
// simpler.
// 1. The first character on the line is decisive. If it is a data item
// it has to be a "_". A loop starts with loop...
// 2. The pdb promises that they will restrict themselves to a certain
// style. In the ATOM records, they always use the same columns and in
// the same order.
// We treat multi-line fields as we should, but this does not seem
// sensible. According to https://www.iucr.org/resources/cif/spec/version1.1/cifsyntax,
// These lines
// ;a
//   b
//;
// Should be read, keeping the newline and space before the b

// Overall structure
// There is a lot of information that will never be of interest to us (solvents,
// crystallisation details, ..). This means that we do a first filtering by
// jumping over parts that are not in our table of interesting data.
// Next, there are tables from which we only want small pieces. The coordinates
// are stored in a table with about 8 columns. We will usually only want
// a name and x, y, z coordinates. In this case, we read the whole table
// and make it available to pick out bits.
// We do not want to keep this table, so we make a structure
//   pdb_raw
// where we put this kind of thing. This must be on the stack or allocated on the
// stack so it goes away when we are finished reading.

// Notes about the mmcif format...
// A question mark, ?, means a missing value.
// A dot, ., means not appropriate or deliberately left out.
// What is what..
// There are entities and chains.
// Entities can be anything - protein, ligands.. There is a mapping back to
// old pdb chains, according to http://mmcif.wwpdb.org/docs/pdb_to_pdbx_correspondences.html, that is called,
// _atom_site.auth_asym_id 
package mmcif
