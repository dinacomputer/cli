package api

import "net/url"

// ---------- Buckets ----------

func (c *Client) ListBuckets(orgID string) ([]Bucket, error) {
	req, err := c.newRequest("GET", "/organizations/"+orgID+"/buckets", nil)
	if err != nil {
		return nil, err
	}
	var out ListBucketsOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Buckets, nil
}

func (c *Client) CreateBucket(orgID string, input CreateBucketInput) (*Bucket, error) {
	req, err := c.newRequest("POST", "/organizations/"+orgID+"/buckets", input)
	if err != nil {
		return nil, err
	}
	var b Bucket
	return &b, c.do(req, &b)
}

func (c *Client) GetBucket(orgID, name string) (*Bucket, error) {
	req, err := c.newRequest("GET", bucketPath(orgID, name), nil)
	if err != nil {
		return nil, err
	}
	var b Bucket
	return &b, c.do(req, &b)
}

func (c *Client) UpdateBucket(orgID, name string, input UpdateBucketInput) (*Bucket, error) {
	req, err := c.newRequest("PATCH", bucketPath(orgID, name), input)
	if err != nil {
		return nil, err
	}
	var b Bucket
	return &b, c.do(req, &b)
}

func (c *Client) DeleteBucket(orgID, name string) error {
	req, err := c.newRequest("DELETE", bucketPath(orgID, name), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// ---------- Storage keys ----------

func (c *Client) ListStorageKeys(orgID string) ([]StorageKey, error) {
	req, err := c.newRequest("GET", "/organizations/"+orgID+"/storage-keys", nil)
	if err != nil {
		return nil, err
	}
	var out ListStorageKeysOutput
	if err := c.do(req, &out); err != nil {
		return nil, err
	}
	return out.Keys, nil
}

func (c *Client) CreateStorageKey(orgID string, input CreateStorageKeyInput) (*IssuedStorageKey, error) {
	req, err := c.newRequest("POST", "/organizations/"+orgID+"/storage-keys", input)
	if err != nil {
		return nil, err
	}
	var issued IssuedStorageKey
	return &issued, c.do(req, &issued)
}

func (c *Client) GetStorageKey(orgID, name string) (*StorageKey, error) {
	req, err := c.newRequest("GET", storageKeyPath(orgID, name), nil)
	if err != nil {
		return nil, err
	}
	var k StorageKey
	return &k, c.do(req, &k)
}

func (c *Client) RotateStorageKey(orgID, name string) (*IssuedStorageKey, error) {
	req, err := c.newRequest("POST", storageKeyPath(orgID, name)+"/rotate", nil)
	if err != nil {
		return nil, err
	}
	var issued IssuedStorageKey
	return &issued, c.do(req, &issued)
}

func (c *Client) DeleteStorageKey(orgID, name string) error {
	req, err := c.newRequest("DELETE", storageKeyPath(orgID, name), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// ---------- Grants ----------

func (c *Client) SetStorageKeyGrant(orgID, keyName, bucket string, input SetGrantInput) (*StorageGrant, error) {
	req, err := c.newRequest("PUT", grantPath(orgID, keyName, bucket), input)
	if err != nil {
		return nil, err
	}
	var g StorageGrant
	return &g, c.do(req, &g)
}

func (c *Client) DeleteStorageKeyGrant(orgID, keyName, bucket string) error {
	req, err := c.newRequest("DELETE", grantPath(orgID, keyName, bucket), nil)
	if err != nil {
		return err
	}
	return c.do(req, nil)
}

// Bucket and key names are user-supplied and travel as path segments, so they
// are escaped rather than concatenated raw.

func bucketPath(orgID, name string) string {
	return "/organizations/" + orgID + "/buckets/" + url.PathEscape(name)
}

func storageKeyPath(orgID, name string) string {
	return "/organizations/" + orgID + "/storage-keys/" + url.PathEscape(name)
}

func grantPath(orgID, keyName, bucket string) string {
	return storageKeyPath(orgID, keyName) + "/grants/" + url.PathEscape(bucket)
}
