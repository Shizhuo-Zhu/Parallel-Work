import React, { useState, useEffect } from 'react';
import axios from 'axios';

const ResourceTable = () => {
  const [resources, setResources] = useState([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [resourceID, setResourceID] = useState(''); // State to hold the resource ID input
  const [region, setRegion] = useState(''); // State to hold the region input
  const [resourceType, setResourceType] = useState(''); // State to hold the type input

  // Function to fetch resources by ID
  const fetchResourceByID = (id) => {
    setLoading(true);
    const apiUrl = `http://localhost:8080/api/resources/${id}`;
    axios.get(apiUrl)
      .then(response => {
        setResources(response.data); // Set the response as the only resource in the list
        setLoading(false);
      })
      .catch(error => {
        setError('Failed to fetch resource by ID');
        setLoading(false);
      });
  };

  // Function to fetch filtered resources based on region and type
  const fetchResourcesByRegionAndType = () => {
    setLoading(true);
    
    let apiUrl = 'http://localhost:8080/api/resources';
    const filters = [];
    if (region) filters.push(`region=${region}`);
    if (resourceType) filters.push(`type=${resourceType}`);

    if (filters.length > 0) {
      apiUrl += '?' + filters.join('&');
    }

    axios.get(apiUrl)
      .then(response => {
        setResources(response.data);
        setLoading(false);
      })
      .catch(error => {
        setError('Failed to fetch resources');
        setLoading(false);
      });
  };

  // Function to fetch all resources
  const fetchAllResources = () => {
    setLoading(true);
    axios.get('http://localhost:8080/api/resources')
      .then(response => {
        setResources(response.data);
        setLoading(false);
      })
      .catch(error => {
        setError('Failed to fetch resources');
        setLoading(false);
      });
  };

  useEffect(() => {
    fetchAllResources(); // Fetch all resources when the component mounts
  }, []);

  const handleIDChange = (event) => {
    setResourceID(event.target.value); // Update the resource ID as the user types
  };

  const handleRegionChange = (event) => {
    setRegion(event.target.value); // Update the region as the user types
  };

  const handleTypeChange = (event) => {
    setResourceType(event.target.value); // Update the type as the user types
  };

  const handleSearchByID = () => {
    if (resourceID) {
      fetchResourceByID(resourceID); // Fetch the resource by ID
    } else {
      setError('Please enter a valid resource ID');
    }
  };

  const handleSearchByRegionAndType = () => {
    fetchResourcesByRegionAndType(); // Fetch resources filtered by region and type
  };

  if (loading) {
    return <div>Loading...</div>;
  }

  if (error) {
    return <div>{error}</div>;
  }

  return (
    <div>
      {/* Input and button to fetch a resource by ID */}
      <div>
        <input
          type="text"
          value={resourceID}
          onChange={handleIDChange}
          placeholder="Enter ID"
        />
        <button onClick={handleSearchByID}>Get Resource by ID</button>
      </div>

      {/* Input fields for region and type */}
      <div>
        <input
          type="text"
          value={region}
          onChange={handleRegionChange}
          placeholder="Enter Region"
        />
        <input
          type="text"
          value={resourceType}
          onChange={handleTypeChange}
          placeholder="Enter Type"
        />
        <button onClick={handleSearchByRegionAndType}>Search</button>
      </div>

      {/* Table to display the resources */}
      <table>
        <thead>
          <tr>
            <th>Name</th>
            <th>Region</th>
            <th>Type</th>
            <th>Status</th>
            <th>IP</th>
            <th>Created Time</th>
          </tr>
        </thead>
        <tbody>
          {resources.map(resource => (
            <tr key={resource.name}>
              <td>{resource.name}</td>
              <td>{resource.zone}</td>
              <td>{resource.type}</td>
              <td>{resource.status || 'N/A'}</td>
              <td>{resource.ipAddresses && resource.ipAddresses.length > 0 ? resource.ipAddresses.join(', ') : 'N/A'}</td>
              <td>{resource.creationTimestamp || 'N/A'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
};

export default ResourceTable;

