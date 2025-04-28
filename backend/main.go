package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"
)

// Custom struct to include additional fields
type Resource struct {
	Name              string   `json:"name"`
	Zone              string   `json:"zone"`
	Type              string   `json:"type"` // "instance" or "disk"
	Status            string   `json:"status"`
	IPAddresses       []string `json:"ipAddresses,omitempty"`
	CreationTimestamp string   `json:"creationTimestamp"`
}

var resourceID string

func main() {
	// Define a command-line argument for resource ID
	flag.StringVar(&resourceID, "id", "", "The ID of the resource to fetch")
	flag.Parse() // Parse the command-line arguments

	r := gin.Default()

	r.Use(cors.Default()) 

	// If the ID is provided via command-line argument, fetch that resource
	if resourceID != "" {
		r.GET("/api/resources", func(c *gin.Context) {
			getResourceByID(c) // Directly call getResourceByID
		})
	} else {
		// Routes for both functions
		r.GET("/api/resources", getResources)
		r.GET("/api/resources/:id", getResourceByID) // This route requires an ID
	}

	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}



// In your backend, update the getResources function to handle region and type filtering
func getResources(c *gin.Context) {
	// Retrieve filters from query parameters for region and type
	regionFilter := c.DefaultQuery("region", "")
	typeFilter := c.DefaultQuery("type", "")

	svc, err := compute.NewService(c, option.WithCredentialsFile("application_default_credentials.json"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to create GCP service: %v", err)})
		return
	}

	projectID := "interns-test-2025"
	zoneList, err := svc.Zones.List(projectID).Do()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to fetch zones: %v", err)})
		return
	}

	var allResources []Resource
	var wg sync.WaitGroup
	var mu sync.Mutex // To safely append to allResources in goroutines

	for _, zone := range zoneList.Items {
		wg.Add(1)
		go func(zone *compute.Zone) {
			defer wg.Done()

			// Fetch instances
			instanceList, err := svc.Instances.List(projectID, zone.Name).Do()
			if err == nil {
				for _, instance := range instanceList.Items {
					if (regionFilter == "" || strings.Contains(instance.Zone, regionFilter)) &&
						(typeFilter == "" || typeFilter == "instance") {
						resource := Resource{
							Name:              instance.Name,
							Zone:              extractZoneName(instance.Zone),
							Type:              "instance",
							Status:            instance.Status,
							IPAddresses:       extractIP(instance.NetworkInterfaces),
							CreationTimestamp: formatCreationTimestamp(instance.CreationTimestamp),
						}
						mu.Lock()
						allResources = append(allResources, resource)
						mu.Unlock()
					}
				}
			}

			// Fetch disks
			diskList, err := svc.Disks.List(projectID, zone.Name).Do()
			if err == nil {
				for _, disk := range diskList.Items {
					if (regionFilter == "" || strings.Contains(disk.Zone, regionFilter)) &&
						(typeFilter == "" || typeFilter == "disk") {
						resource := Resource{
							Name:              disk.Name,
							Zone:              extractZoneName(disk.Zone),
							Type:              "disk",
							Status:            disk.Status,
							CreationTimestamp: formatCreationTimestamp(disk.CreationTimestamp),
						}
						mu.Lock()
						allResources = append(allResources, resource)
						mu.Unlock()
					}
				}
			}
		}(zone)
	}

	wg.Wait()

	c.JSON(http.StatusOK, allResources)
}






// func getResourceByID(c *gin.Context) {
// 	resourceIDParam := c.Param("id")
// 	// Use the command-line argument if the ID is passed from the command line
// 	if resourceID == "" {
// 		resourceID = resourceIDParam
// 	}

// 	if resourceID == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID is required"})
// 		return
// 	}

// 	svc, err := compute.NewService(c, option.WithCredentialsFile("application_default_credentials.json"))
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to create GCP service: %v", err)})
// 		return
// 	}

// 	projectID := "interns-test-2025"
// 	zoneList, err := svc.Zones.List(projectID).Do()
// 	if err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to fetch zones: %v", err)})
// 		return
// 	}

// 	// First try to find the instance
// 	for _, zone := range zoneList.Items {
// 		instance, err := svc.Instances.Get(projectID, zone.Name, resourceID).Do()
// 		if err == nil {
// 			resource := Resource{
// 				Name:              instance.Name,
// 				Zone:              extractZoneName(instance.Zone),
// 				Type:              "instance",
// 				Status:            instance.Status,
// 				IPAddresses:       extractIP(instance.NetworkInterfaces),
// 				CreationTimestamp: formatCreationTimestamp(instance.CreationTimestamp),
// 			}
// 			c.JSON(http.StatusOK, resource)
// 			return
// 		}
// 	}

// 	// Then try to find the disk
// 	for _, zone := range zoneList.Items {
// 		disk, err := svc.Disks.Get(projectID, zone.Name, resourceID).Do()
// 		if err == nil {
// 			resource := Resource{
// 				Name:              disk.Name,
// 				Zone:              extractZoneName(disk.Zone),
// 				Type:              "disk",
// 				Status:            disk.Status,
// 				CreationTimestamp: formatCreationTimestamp(disk.CreationTimestamp),
// 			}
// 			c.JSON(http.StatusOK, resource)
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
// }

func getResourceByID(c *gin.Context) {
    resourceIDParam := c.Param("id")
    // Use the command-line argument if the ID is passed from the command line
    if resourceID == "" {
        resourceID = resourceIDParam
    }

    if resourceID == "" {
        c.JSON(http.StatusBadRequest, gin.H{"error": "Resource ID is required"})
        return
    }

    svc, err := compute.NewService(c, option.WithCredentialsFile("application_default_credentials.json"))
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to create GCP service: %v", err)})
        return
    }

    projectID := "interns-test-2025"
    zoneList, err := svc.Zones.List(projectID).Do()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Unable to fetch zones: %v", err)})
        return
    }

    // First try to find the instance
    for _, zone := range zoneList.Items {
        instance, err := svc.Instances.Get(projectID, zone.Name, resourceID).Do()
        if err == nil {
            resource := Resource{
                Name:              instance.Name,
                Zone:              extractZoneName(instance.Zone),
                Type:              "instance",
                Status:            instance.Status,
                IPAddresses:       extractIP(instance.NetworkInterfaces),
                CreationTimestamp: formatCreationTimestamp(instance.CreationTimestamp),
            }
            // Return the resource as an array
            c.JSON(http.StatusOK, []Resource{resource})
            return
        }
    }

    // Then try to find the disk
    for _, zone := range zoneList.Items {
        disk, err := svc.Disks.Get(projectID, zone.Name, resourceID).Do()
        if err == nil {
            resource := Resource{
                Name:              disk.Name,
                Zone:              extractZoneName(disk.Zone),
                Type:              "disk",
                Status:            disk.Status,
                CreationTimestamp: formatCreationTimestamp(disk.CreationTimestamp),
            }
            // Return the resource as an array
            c.JSON(http.StatusOK, []Resource{resource})
            return
        }
    }

    c.JSON(http.StatusNotFound, gin.H{"error": "Resource not found"})
}


// Utility functions
func extractZoneName(zoneURL string) string {
	parts := strings.Split(zoneURL, "/")
	return parts[len(parts)-1]
}

func extractIP(networkInterfaces []*compute.NetworkInterface) []string {
	var ipAddresses []string
	for _, ni := range networkInterfaces {
		if len(ni.AccessConfigs) > 0 {
			for _, ac := range ni.AccessConfigs {
				if ac.NatIP != "" {
					ipAddresses = append(ipAddresses, ac.NatIP)
				}
			}
		}
	}
	return ipAddresses
}

func formatCreationTimestamp(timestamp string) string {
	parsedTime, err := time.Parse(time.RFC3339, timestamp)
	if err != nil {
		log.Printf("Error parsing creation timestamp: %v", err)
		return "N/A"
	}
	return parsedTime.Format("2006-01-02")
}
