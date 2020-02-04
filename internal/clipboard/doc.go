// Copyright © 2020 Phil Pennock.
// All rights reserved, except as granted under license.
// Licensed per file LICENSE.txt

/*
Package clipboard is an internal shim around the external clipboard package
which we use, to enable building without that package for environments where
it's not available (eg, js/wasm).
*/
package clipboard
