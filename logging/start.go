package logging

import (
  "fmt"
  "bytes"
  "strings"
  "golang.org/x/net/context"
  "github.com/docker/docker/client"
  "github.com/docker/docker/api/types"
  "github.com/docker/docker/api/types/events"
  "github.com/docker/docker/api/types/container"
  networktypes "github.com/docker/docker/api/types/network"
  //"reflect"
)

func StartLogging(client *client.Client, message events.Message) {
  var buffer bytes.Buffer
  basecontainer := message.Actor.Attributes["name"]
  buffer.WriteString(basecontainer)
  buffer.WriteString("_filebeat")
  containername := buffer.String()
  inspect_result, err := client.ContainerInspect(context.Background(), containername)
  if err != nil  {
    if strings.Contains(err.Error(), "No such container:") {
      fmt.Println("Logging container doesn't exist lets create and start it")
      //fmt.Println(message.Actor.Attributes["application-type"])
      labels := make(map[string]string)
      labels["origin"] = "datamgmt"
      labels["application-type"] = "filebeat"
      //test := network.EndpointSettings{ NetworkID: "fd9af838d33f" }
      /*network := map[string]interface{}{
		    "NetworkingConfig": networktypes.NetworkingConfig{
			    EndpointsConfig: map[string]*networktypes.EndpointSettings{
				    "datamgmt": {},
			    },
		    },
	    }*/
      network := map[string]*networktypes.EndpointSettings{
			    "datamgmt": {},
		  }
      r, err := client.ContainerCreate(context.Background(), &container.Config{Hostname: containername, Labels: labels, Image: "cloudunit/datamgmt-filebeat:latest", Env: []string{"APPLICATION_TYPE="+message.Actor.Attributes["application-type"]}}, &container.HostConfig{ VolumesFrom: []string{basecontainer}}, &networktypes.NetworkingConfig{ EndpointsConfig: network}, containername)
      if err != nil {
        fmt.Println("Could not create filebeat container", err)
      } else {
        err = client.ContainerStart(context.Background(), r.ID, types.ContainerStartOptions{})
        if err != nil {
          fmt.Println("Cannot start filebeat container", err)
        }
      }
    } else {
      fmt.Println("Error on container inspect", err)
    }
  } else {
    if inspect_result.State.Status != "running" {
      fmt.Println("Lets start logging container")
      err := client.ContainerStart(context.Background(), inspect_result.ID, types.ContainerStartOptions{})
      if err != nil {
        fmt.Println("Cannot start filebeat container associated to ", basecontainer, err)
      }
    } else {
      fmt.Println("Container aleady running")
    }
  }
}
