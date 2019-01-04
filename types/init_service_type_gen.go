package types

// Code generated by github.com/CovenantSQL/HashStablePack DO NOT EDIT.

import (
	hsp "github.com/CovenantSQL/HashStablePack/marshalhash"
)

// MarshalHash marshals for hash
func (z *InitService) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 1
	o = append(o, 0x81, 0x81)
	if oTemp, err := z.Envelope.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *InitService) Msgsize() (s int) {
	s = 1 + 9 + z.Envelope.Msgsize()
	return
}

// MarshalHash marshals for hash
func (z *InitServiceResponse) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 1
	o = append(o, 0x81, 0x81)
	if oTemp, err := z.Header.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *InitServiceResponse) Msgsize() (s int) {
	s = 1 + 7 + z.Header.Msgsize()
	return
}

// MarshalHash marshals for hash
func (z *InitServiceResponseHeader) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 1
	o = append(o, 0x81, 0x81)
	o = hsp.AppendArrayHeader(o, uint32(len(z.Instances)))
	for za0001 := range z.Instances {
		if oTemp, err := z.Instances[za0001].MarshalHash(); err != nil {
			return nil, err
		} else {
			o = hsp.AppendBytes(o, oTemp)
		}
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *InitServiceResponseHeader) Msgsize() (s int) {
	s = 1 + 10 + hsp.ArrayHeaderSize
	for za0001 := range z.Instances {
		s += z.Instances[za0001].Msgsize()
	}
	return
}

// MarshalHash marshals for hash
func (z *ResourceMeta) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 8
	o = append(o, 0x88, 0x88)
	o = hsp.AppendArrayHeader(o, uint32(len(z.TargetMiners)))
	for za0001 := range z.TargetMiners {
		if oTemp, err := z.TargetMiners[za0001].MarshalHash(); err != nil {
			return nil, err
		} else {
			o = hsp.AppendBytes(o, oTemp)
		}
	}
	o = append(o, 0x88)
	o = hsp.AppendBool(o, z.UseEventualConsistency)
	o = append(o, 0x88)
	o = hsp.AppendFloat64(o, z.ConsistencyLevel)
	o = append(o, 0x88)
	o = hsp.AppendFloat64(o, z.LoadAvgPerCPU)
	o = append(o, 0x88)
	o = hsp.AppendString(o, z.EncryptionKey)
	o = append(o, 0x88)
	o = hsp.AppendUint16(o, z.Node)
	o = append(o, 0x88)
	o = hsp.AppendUint64(o, z.Space)
	o = append(o, 0x88)
	o = hsp.AppendUint64(o, z.Memory)
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ResourceMeta) Msgsize() (s int) {
	s = 1 + 13 + hsp.ArrayHeaderSize
	for za0001 := range z.TargetMiners {
		s += z.TargetMiners[za0001].Msgsize()
	}
	s += 23 + hsp.BoolSize + 17 + hsp.Float64Size + 14 + hsp.Float64Size + 14 + hsp.StringPrefixSize + len(z.EncryptionKey) + 5 + hsp.Uint16Size + 6 + hsp.Uint64Size + 7 + hsp.Uint64Size
	return
}

// MarshalHash marshals for hash
func (z *ServiceInstance) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 4
	o = append(o, 0x84, 0x84)
	if z.GenesisBlock == nil {
		o = hsp.AppendNil(o)
	} else {
		if oTemp, err := z.GenesisBlock.MarshalHash(); err != nil {
			return nil, err
		} else {
			o = hsp.AppendBytes(o, oTemp)
		}
	}
	o = append(o, 0x84)
	if z.Peers == nil {
		o = hsp.AppendNil(o)
	} else {
		if oTemp, err := z.Peers.MarshalHash(); err != nil {
			return nil, err
		} else {
			o = hsp.AppendBytes(o, oTemp)
		}
	}
	o = append(o, 0x84)
	if oTemp, err := z.ResourceMeta.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	o = append(o, 0x84)
	if oTemp, err := z.DatabaseID.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *ServiceInstance) Msgsize() (s int) {
	s = 1 + 13
	if z.GenesisBlock == nil {
		s += hsp.NilSize
	} else {
		s += z.GenesisBlock.Msgsize()
	}
	s += 6
	if z.Peers == nil {
		s += hsp.NilSize
	} else {
		s += z.Peers.Msgsize()
	}
	s += 13 + z.ResourceMeta.Msgsize() + 11 + z.DatabaseID.Msgsize()
	return
}

// MarshalHash marshals for hash
func (z *SignedInitServiceResponseHeader) MarshalHash() (o []byte, err error) {
	var b []byte
	o = hsp.Require(b, z.Msgsize())
	// map header, size 2
	// map header, size 1
	o = append(o, 0x82, 0x82, 0x81, 0x81)
	o = hsp.AppendArrayHeader(o, uint32(len(z.InitServiceResponseHeader.Instances)))
	for za0001 := range z.InitServiceResponseHeader.Instances {
		if oTemp, err := z.InitServiceResponseHeader.Instances[za0001].MarshalHash(); err != nil {
			return nil, err
		} else {
			o = hsp.AppendBytes(o, oTemp)
		}
	}
	o = append(o, 0x82)
	if oTemp, err := z.DefaultHashSignVerifierImpl.MarshalHash(); err != nil {
		return nil, err
	} else {
		o = hsp.AppendBytes(o, oTemp)
	}
	return
}

// Msgsize returns an upper bound estimate of the number of bytes occupied by the serialized message
func (z *SignedInitServiceResponseHeader) Msgsize() (s int) {
	s = 1 + 26 + 1 + 10 + hsp.ArrayHeaderSize
	for za0001 := range z.InitServiceResponseHeader.Instances {
		s += z.InitServiceResponseHeader.Instances[za0001].Msgsize()
	}
	s += 28 + z.DefaultHashSignVerifierImpl.Msgsize()
	return
}
