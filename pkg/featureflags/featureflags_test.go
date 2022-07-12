package featureflags

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("featureflags", func() {
	BeforeEach(func() {
		// Clear flags
		flags = make(map[string]string)
	})

	It("updates when Set is called", func() {
		Expect(flags).To(BeEmpty())

		Set("FLAG", "is set")
		Expect(flags).To(HaveKeyWithValue("FLAG", "is set"))

		Set("OTHER_FLAG", "is also set")
		Expect(flags).To(HaveKeyWithValue("OTHER_FLAG", "is also set"))

		Set("FLAG", "other value")
		Expect(flags).To(HaveKeyWithValue("FLAG", "other value"))
	})

	It("returns flags when Get is called", func() {
		Expect(Get("FLAG")).To(Equal(""))

		Set("FLAG", "is set")
		Expect(Get("FLAG")).To(Equal("is set"))
	})

	It("returns all flags when GetFlags is called", func() {
		Expect(GetFlags()).To(BeEmpty())

		Set("FLAG", "is set")
		Expect(GetFlags()).To(HaveKeyWithValue("FLAG", "is set"))
		Expect(GetFlags()).To(HaveLen(1))

		Set("FLAG", "other value")
		Expect(GetFlags()).To(HaveKeyWithValue("FLAG", "other value"))
		Expect(GetFlags()).To(HaveLen(1))

		Set("OTHER_FLAG", "some value")
		Expect(GetFlags()).To(HaveKeyWithValue("FLAG", "other value"))
		Expect(GetFlags()).To(HaveKeyWithValue("OTHER_FLAG", "some value"))
		Expect(GetFlags()).To(HaveLen(2))

		Get("Key that doesn't exist")
		Expect(GetFlags()).To(HaveKeyWithValue("FLAG", "other value"))
		Expect(GetFlags()).To(HaveKeyWithValue("OTHER_FLAG", "some value"))
		Expect(GetFlags()).To(HaveLen(2))
	})
})
