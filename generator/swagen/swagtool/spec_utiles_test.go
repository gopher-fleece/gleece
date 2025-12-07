package swagtool

import (
	"github.com/gopher-fleece/runtime"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("Swagtools - Spec Utilities", func() {

	Describe("HttpStatusCodeToString", func() {
		It("should convert HttpStatusCode to string", func() {
			code := runtime.HttpStatusCode(200)
			Expect(HttpStatusCodeToString(code)).To(Equal("200"))
		})
	})

	Describe("ParseNumber", func() {
		It("should parse a valid number string", func() {
			Expect(*ParseNumber("123.45")).To((BeEquivalentTo(123.45)))
		})

		It("should return nil for an invalid number string", func() {
			Expect(ParseNumber("abc")).To(BeNil())
		})
	})

	Describe("ParseInteger", func() {
		It("should parse a valid integer string", func() {
			Expect(*ParseInteger("123")).To((BeEquivalentTo(123)))
		})

		It("should return nil for an invalid integer string", func() {
			Expect(ParseInteger("abc")).To(BeNil())
		})
	})

	Describe("ParseUInteger", func() {
		It("should parse a valid integer string", func() {
			Expect(*ParseUInteger("123")).To((BeEquivalentTo(123)))
		})

		It("should return nil for an invalid integer string", func() {
			Expect(ParseInteger("abc")).To(BeNil())
		})
	})

	Describe("ParseBool", func() {
		It("should parse a valid boolean string", func() {
			Expect(*ParseBool("true")).To((BeTrue()))
			Expect(*ParseBool("false")).To((BeFalse()))
		})

		It("should return nil for an invalid boolean string", func() {
			Expect(ParseBool("notabool")).To(BeNil())
		})
	})

	Describe("ForceOrderedJSON", func() {
		It("should order JSON keys alphabetically", func() {
			// Unordered JSON input
			input := []byte(`{"zebra":"last","apple":"first","middle":"second"}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			// Expected output with keys ordered alphabetically
			expected := `{
  "apple": "first",
  "middle": "second",
  "zebra": "last"
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should order nested JSON objects", func() {
			input := []byte(`{"outer":{"zebra":"z","apple":"a"},"first":"value"}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "first": "value",
  "outer": {
    "apple": "a",
    "zebra": "z"
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should handle arrays in JSON", func() {
			input := []byte(`{"items":[3,1,2],"name":"test"}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			// Arrays should maintain order, only keys should be sorted
			expected := `{
  "items": [
    3,
    1,
    2
  ],
  "name": "test"
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should produce deterministic output for same input", func() {
			input := []byte(`{"z":"last","a":"first","m":"middle"}`)

			result1, err1 := ForceOrderedJSON(input)
			Expect(err1).To(BeNil())

			result2, err2 := ForceOrderedJSON(input)
			Expect(err2).To(BeNil())

			// Both results should be identical
			Expect(string(result1)).To(Equal(string(result2)))
		})

		It("should handle complex OpenAPI-like structures", func() {
			input := []byte(`{
				"paths": {"/users": {"get": {}}},
				"openapi": "3.1.0",
				"info": {"title": "API"}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "info": {
    "title": "API"
  },
  "openapi": "3.1.0",
  "paths": {
    "/users": {
      "get": {}
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should return error for invalid JSON", func() {
			input := []byte(`{invalid json}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).NotTo(BeNil())
			Expect(result).To(BeNil())
			Expect(err.Error()).To(ContainSubstring("error unmarshaling JSON"))
		})

		It("should handle empty JSON object", func() {
			input := []byte(`{}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())
			Expect(string(result)).To(Equal("{}"))
		})

		It("should handle JSON with different value types", func() {
			input := []byte(`{"string":"text","number":42,"bool":true,"null":null}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "bool": true,
  "null": null,
  "number": 42,
  "string": "text"
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should sort enum values in components.schemas", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"Status": {
							"type": "string",
							"enum": ["pending", "active", "completed", "archived"]
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "Status": {
        "enum": [
          "active",
          "archived",
          "completed",
          "pending"
        ],
        "type": "string"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should sort numeric enum values", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"Priority": {
							"type": "integer",
							"enum": [3, 1, 5, 2, 4]
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "Priority": {
        "enum": [
          1,
          2,
          3,
          4,
          5
        ],
        "type": "integer"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should sort enum values in nested properties", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"User": {
							"type": "object",
							"properties": {
								"status": {
									"type": "string",
									"enum": ["inactive", "active", "banned"]
								},
								"role": {
									"type": "string",
									"enum": ["user", "admin", "guest"]
								}
							}
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "User": {
        "properties": {
          "role": {
            "enum": [
              "admin",
              "guest",
              "user"
            ],
            "type": "string"
          },
          "status": {
            "enum": [
              "active",
              "banned",
              "inactive"
            ],
            "type": "string"
          }
        },
        "type": "object"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should sort enum values in array items", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"StatusArray": {
							"type": "array",
							"items": {
								"type": "string",
								"enum": ["draft", "published", "archived"]
							}
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "StatusArray": {
        "items": {
          "enum": [
            "archived",
            "draft",
            "published"
          ],
          "type": "string"
        },
        "type": "array"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should sort enum values in allOf, oneOf, anyOf", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"Combined": {
							"oneOf": [
								{
									"type": "string",
									"enum": ["z", "a", "m"]
								}
							],
							"allOf": [
								{
									"type": "string",
									"enum": ["3", "1", "2"]
								}
							],
							"anyOf": [
								{
									"type": "string",
									"enum": ["y", "x", "z"]
								}
							]
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "Combined": {
        "allOf": [
          {
            "enum": [
              "1",
              "2",
              "3"
            ],
            "type": "string"
          }
        ],
        "anyOf": [
          {
            "enum": [
              "x",
              "y",
              "z"
            ],
            "type": "string"
          }
        ],
        "oneOf": [
          {
            "enum": [
              "a",
              "m",
              "z"
            ],
            "type": "string"
          }
        ]
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should sort enum values in additionalProperties", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"MapType": {
							"type": "object",
							"additionalProperties": {
								"type": "string",
								"enum": ["value3", "value1", "value2"]
							}
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "MapType": {
        "additionalProperties": {
          "enum": [
            "value1",
            "value2",
            "value3"
          ],
          "type": "string"
        },
        "type": "object"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should handle empty enum arrays", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"Empty": {
							"type": "string",
							"enum": []
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "Empty": {
        "enum": [],
        "type": "string"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should handle single enum value", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"Single": {
							"type": "string",
							"enum": ["only"]
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "Single": {
        "enum": [
          "only"
        ],
        "type": "string"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should not sort non-enum arrays", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"NotEnum": {
							"type": "array",
							"items": {"type": "string"},
							"example": ["z", "a", "m"]
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			// example array should maintain original order
			expected := `{
  "components": {
    "schemas": {
      "NotEnum": {
        "example": [
          "z",
          "a",
          "m"
        ],
        "items": {
          "type": "string"
        },
        "type": "array"
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should handle JSON without components.schemas", func() {
			input := []byte(`{
				"openapi": "3.1.0",
				"info": {
					"title": "API"
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "info": {
    "title": "API"
  },
  "openapi": "3.1.0"
}`
			Expect(string(result)).To(Equal(expected))
		})

		It("should handle multiple schemas with enum values", func() {
			input := []byte(`{
				"components": {
					"schemas": {
						"Status": {
							"enum": ["pending", "active"]
						},
						"Role": {
							"enum": ["user", "admin"]
						}
					}
				}
			}`)

			result, err := ForceOrderedJSON(input)
			Expect(err).To(BeNil())

			expected := `{
  "components": {
    "schemas": {
      "Role": {
        "enum": [
          "admin",
          "user"
        ]
      },
      "Status": {
        "enum": [
          "active",
          "pending"
        ]
      }
    }
  }
}`
			Expect(string(result)).To(Equal(expected))
		})
	})
})
