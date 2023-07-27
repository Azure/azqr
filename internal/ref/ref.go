// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package ref

func Of[E any](e E) *E {
	return &e
}
