package common

func CheckerUnmarshalSeal(c *ChainChecker) error {
	var message []byte
	if err := c.ContextValue("message", &message); err != nil {
		return err
	}

	var networkID NetworkID
	if err := c.ContextValue("networkID", &networkID); err != nil {
		return err
	}

	var sealCodec *SealCodec
	if err := c.ContextValue("sealCodec", &sealCodec); err != nil {
		return err
	}

	seal, err := sealCodec.Decode(message)
	if err != nil {
		return err
	}

	if err := seal.Wellformed(); err != nil {
		return err
	} else if err := seal.CheckSignature(networkID); err != nil {
		return err
	}

	_ = c.SetContext("seal", seal)

	return nil
}
