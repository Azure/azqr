// Copyright (c) Microsoft Corporation.
// Licensed under the MIT License.

package to

func Ptr[E any](e E) *E {
	return &e
}
