package itf

import (
	"errors"
	"fmt"
	"github.com/mdzio/go-hmccu/binrpc"
	"github.com/mdzio/go-hmccu/model"
	"strings"

	"github.com/mdzio/go-hmccu/xmlrpc"

	"github.com/mdzio/go-logging"
)

var clnLog = logging.Get("itf-client")

type rpcClient interface {
	Call(method string, params []*model.Value) (*model.Value, error)
}

// Client provides access to the HomeMatic XML-RPC API.
type Client struct {
	Addr string
	rpcClient
}

// NewClient creates a new Client.
func NewClient(address string) *Client {
	if strings.HasPrefix(address, "http://") {
		return &Client{address, &xmlrpc.Client{Addr: address}}
	}
	return &Client{address, &binrpc.Client{Addr: address}}
}

// GetDeviceDescription retrieves the device description for the specified
// device.
func (c *Client) GetDeviceDescription(deviceAddress string) (*DeviceDescription, error) {
	clnLog.Debugf("Calling method getDeviceDescription(%s) on %s", deviceAddress, c.Addr)
	// execute call
	v, err := c.Call("getDeviceDescription", []*model.Value{
		&model.Value{FlatString: deviceAddress},
	})
	if err != nil {
		return nil, err
	}

	// build result
	e := model.Q(v)
	d := &DeviceDescription{}
	d.ReadFrom(e)
	if e.Err() != nil {
		return nil, fmt.Errorf("Invalid XML response for getDeviceDescription: %v", e.Err())
	}
	return d, nil
}

// ListDevices retrieves the device descriptions from all devices.
func (c *Client) ListDevices() ([]*DeviceDescription, error) {
	clnLog.Debugf("Calling method listDevices on %s", c.Addr)
	// execute call
	v, err := c.Call("listDevices", []*model.Value{})
	if err != nil {
		return nil, err
	}

	// build result
	e := model.Q(v)
	var r []*DeviceDescription
	for _, av := range e.Slice() {
		d := &DeviceDescription{}
		d.ReadFrom(av)
		r = append(r, d)
	}

	if e.Err() != nil {
		return nil, fmt.Errorf("Invalid XML response for listDevices: %v", e.Err())
	}
	return r, nil
}

// GetParamsetDescription retrieves the paramset description from a device.
func (c *Client) GetParamsetDescription(deviceAddress string, paramsetType string) (ParamsetDescription, error) {
	clnLog.Debugf("Calling method getParamsetDescription(%s, %s) on %s", deviceAddress, paramsetType, c.Addr)
	// execute call
	v, err := c.Call("getParamsetDescription", []*model.Value{
		&model.Value{FlatString: deviceAddress},
		&model.Value{FlatString: paramsetType},
	})
	if err != nil {
		return nil, err
	}

	// build result
	e := model.Q(v)
	r := make(ParamsetDescription)
	for n, v := range e.Map() {
		p := &ParameterDescription{}
		p.ReadFrom(v)
		if e.Err() != nil {
			break
		}
		r[n] = p
	}
	if e.Err() != nil {
		return nil, fmt.Errorf("Invalid XML response for getParamsetDescription: %v", e.Err())
	}
	return r, nil
}

// GetParamset retrieves the specified parameter set.
func (c *Client) GetParamset(deviceAddress string, paramsetType string) (map[string]interface{}, error) {
	clnLog.Debugf("Calling method getParamset(%s, %s) on %s", deviceAddress, paramsetType, c.Addr)
	// execute call
	v, err := c.Call("getParamset", []*model.Value{
		&model.Value{FlatString: deviceAddress},
		&model.Value{FlatString: paramsetType},
	})
	if err != nil {
		return nil, err
	}

	// build result
	e := model.Q(v)
	r := make(map[string]interface{})
	for n, v := range e.Map() {
		vv := v.Any()
		if e.Err() != nil {
			break
		}
		r[n] = vv
	}
	if e.Err() != nil {
		return nil, fmt.Errorf("Invalid XML response for getParamset: %v", e.Err())
	}
	return r, nil
}

// PutParamset writes the parameter set.
func (c *Client) PutParamset(deviceAddress string, paramsetType string, paramset map[string]interface{}) error {
	clnLog.Debugf("Calling method putParamset(%s, %s) on %s", deviceAddress, paramsetType, c.Addr)
	// convert value
	ps, err := model.NewValue(paramset)
	if err != nil {
		return err
	}
	// execute call
	resp, err := c.Call("putParamset", []*model.Value{
		&model.Value{FlatString: deviceAddress},
		&model.Value{FlatString: paramsetType},
		ps,
	})
	if err != nil {
		return err
	}
	// assert empty response
	err = c.assertEmptyResponse(resp)
	if err != nil {
		return fmt.Errorf("Invalid response for method putParamset: %v", err)
	}
	return err
}

func (c *Client) assertEmptyResponse(v *model.Value) error {
	eval := model.Q(v)
	s := eval.String()
	if eval.Err() != nil || s != "" {
		return errors.New("Expected emtpy string as response")
	}
	return nil
}

// SetValue sets a single value from the parameter set VALUES.
func (c *Client) SetValue(deviceAddress string, valueName string, value interface{}) error {
	clnLog.Debugf("Calling method setValue(%s, %s, %v) on %s", deviceAddress, valueName, value, c.Addr)
	// convert value
	v, err := model.NewValue(value)
	if err != nil {
		return err
	}
	// execute call
	resp, err := c.Call("setValue", []*model.Value{
		&model.Value{FlatString: deviceAddress},
		&model.Value{FlatString: valueName},
		v,
	})
	if err != nil {
		return err
	}
	// assert empty response
	err = c.assertEmptyResponse(resp)
	if err != nil {
		return fmt.Errorf("Invalid response for method setValue: %v", err)
	}
	return nil
}

// GetValue gets a single value from the parameter set VALUES.
func (c *Client) GetValue(deviceAddress string, valueName string) (interface{}, error) {
	clnLog.Debugf("Calling method getValue(%s, %s) on %s", deviceAddress, valueName, c.Addr)
	// execute call
	resp, err := c.Call("getValue", []*model.Value{
		&model.Value{FlatString: deviceAddress},
		&model.Value{FlatString: valueName},
	})
	if err != nil {
		return nil, err
	}
	// convert response
	q := model.Q(resp)
	res := q.Any()
	if q.Err() != nil {
		return nil, fmt.Errorf("Invalid response from method getValue: %v", q.Err())
	}
	return res, nil
}

// Init registers a new interface. The receiverAddress should have the format
// http://hostname[:port][/Path]. If the path is not specified, the CCU will use
// /RPC2.
func (c *Client) Init(receiverAddress, id string) error {
	clnLog.Debugf("Calling method init(%s, %s) on %s", receiverAddress, id, c.Addr)
	// execute call
	resp, err := c.Call("init", []*model.Value{
		&model.Value{FlatString: receiverAddress},
		&model.Value{FlatString: id},
	})
	if err != nil {
		return err
	}
	// assert empty response
	err = c.assertEmptyResponse(resp)
	if err != nil {
		return fmt.Errorf("Invalid response for method init: %v", err)
	}
	return nil
}

// Deinit deregisters an interface. The receiverAddress should match with Init.
func (c *Client) Deinit(receiverAddress string) error {
	clnLog.Debugf("Calling method init(%s) on %s", receiverAddress, c.Addr)
	// execute call
	resp, err := c.Call("init", []*model.Value{
		&model.Value{FlatString: receiverAddress},
		// omit 2nd parameter
	})
	if err != nil {
		return err
	}
	// assert empty response
	err = c.assertEmptyResponse(resp)
	if err != nil {
		return fmt.Errorf("Invalid response for method init: %v", err)
	}
	return nil
}

// Ping triggers a pong event. Returns true on success.
func (c *Client) Ping(callerID string) (bool, error) {
	clnLog.Debugf("Calling method ping(%s) on %s", callerID, c.Addr)
	// execute call
	resp, err := c.Call("ping", []*model.Value{
		&model.Value{FlatString: callerID},
	})
	if err != nil {
		return false, err
	}
	// bool response
	q := model.Q(resp)
	res := q.Bool()
	if q.Err() != nil {
		// BidCos-RF returns an array with one bool
		q2 := model.Q(resp)
		res = q2.Idx(0).Bool()
		if q2.Err() != nil {
			return false, fmt.Errorf("Invalid response from method ping: %v, %v", q.Err(), q2.Err())
		}
	}
	return res, nil
}

// Event sends an event.
func (c *Client) Event(interfaceID, address, valueKey string, value interface{}) error {
	clnLog.Debugf("Calling method event(%s, %s, %s, %v) on %s", interfaceID, address, valueKey, value, c.Addr)
	// execute call
	v, err := model.NewValue(value)
	if err != nil {
		return err
	}
	resp, err := c.Call("event", []*model.Value{
		&model.Value{FlatString: interfaceID},
		&model.Value{FlatString: address},
		&model.Value{FlatString: valueKey},
		v,
	})
	if err != nil {
		return err
	}
	// assert empty response
	err = c.assertEmptyResponse(resp)
	if err != nil {
		return fmt.Errorf("Invalid response for method event: %v", err)
	}
	return nil
}
