package api

// ---------- Clusters ----------

func (c *Client) ListClusters() ([]Cluster, error) {
	req, err := c.newRequest("GET", "/clusters", nil)
	if err != nil {
		return nil, err
	}
	var out ListClustersOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Clusters, nil
}

func (c *Client) GetCluster(name string) (*Cluster, error) {
	req, err := c.newRequest("GET", "/clusters/"+name, nil)
	if err != nil {
		return nil, err
	}
	var cluster Cluster
	return &cluster, c.do(req, &cluster)
}
