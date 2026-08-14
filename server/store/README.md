# Store Dev Guide

Stores share a common interface in [](interface.go).

Implementations of this interface share a test suite in [](store_contract_test.go) to verify stores act the same.

Serialization methods also live in `store` in an effort to keep the core data model and related methods clean.
