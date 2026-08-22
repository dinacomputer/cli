package api

// ---------- Organizations ----------

func (c *Client) ListOrganizations() ([]Organization, error) {
	req, err := c.newRequest("GET", "/organizations", nil)
	if err != nil {
		return nil, err
	}
	var out ListOrganizationsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Organizations, nil
}

// ---------- Federations ----------

func (c *Client) ListFederations(orgID string) ([]Federation, error) {
	req, err := c.newRequest("GET", "/organizations/"+orgID+"/federations", nil)
	if err != nil {
		return nil, err
	}
	var out ListFederationsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Federations, nil
}

func (c *Client) CreateFederation(orgID string, input CreateFederationInput) (*Federation, error) {
	req, err := c.newRequest("POST", "/organizations/"+orgID+"/federations", input)
	if err != nil {
		return nil, err
	}
	var fed Federation
	return &fed, c.do(req, &fed)
}

func (c *Client) GetFederation(orgID, id string) (*Federation, error) {
	req, err := c.newRequest("GET", "/organizations/"+orgID+"/federations/"+id, nil)
	if err != nil {
		return nil, err
	}
	var fed Federation
	return &fed, c.do(req, &fed)
}

func (c *Client) UpdateFederation(orgID, id string, input UpdateFederationInput) (*Federation, error) {
	req, err := c.newRequest("PATCH", "/organizations/"+orgID+"/federations/"+id, input)
	if err != nil {
		return nil, err
	}
	var fed Federation
	return &fed, c.do(req, &fed)
}

func (c *Client) DeleteFederation(orgID, id string) error {
	req, err := c.newRequest("DELETE", "/organizations/"+orgID+"/federations/"+id, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// ---------- Mappings ----------

func (c *Client) ListMappings(orgID, fedID string) ([]Mapping, error) {
	req, err := c.newRequest("GET", "/organizations/"+orgID+"/federations/"+fedID+"/mappings", nil)
	if err != nil {
		return nil, err
	}
	var out ListMappingsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Mappings, nil
}

func (c *Client) CreateMapping(orgID, fedID string, input CreateMappingInput) (*Mapping, error) {
	req, err := c.newRequest("POST", "/organizations/"+orgID+"/federations/"+fedID+"/mappings", input)
	if err != nil {
		return nil, err
	}
	var m Mapping
	return &m, c.do(req, &m)
}

func (c *Client) DeleteMapping(orgID, fedID, mappingID string) error {
	req, err := c.newRequest("DELETE", "/organizations/"+orgID+"/federations/"+fedID+"/mappings/"+mappingID, nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}
