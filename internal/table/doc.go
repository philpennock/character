// Copyright © 2016 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

/*
Package table is the internal shim for providing table support for laying out
content prettily.

All interaction with the dependency which provides terminal-drawn tables
should go through this module.  This provides a shim, giving us isolation.
We know the exact subset of features which we rely upon, and can switch
providers.  If desired, we can use multiple files with build-tags, to let the
dependency be satisfied at build-time.
*/
package table
