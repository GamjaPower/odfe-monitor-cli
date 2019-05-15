package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	mapset "github.com/deckarep/golang-set"
)

func getRemoteMonitors() (map[string]Monitor, mapset.Set) {
	var (
		r                    map[string]interface{}
		allMonitors          []Monitor
		allRemoteMonitorsMap map[string]Monitor
	)
	byt := []byte(`{"query":{ "match_all": {}}}`)
	resp, err := http.Post("http://localhost:9200/_opendistro/_alerting/monitors/_search", "application/json", bytes.NewBuffer(byt))
	if err != nil {
		fmt.Println("Error retriving all the monitors", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&r)
	// Print the ID and document source for each hit.
	for _, hit := range r["hits"].(map[string]interface{})["hits"].([]interface{}) {
		var monitor Monitor
		parsedMonitor, err := json.Marshal(hit.(map[string]interface{})["_source"])
		if err != nil {
			fmt.Println("invalid json in the monitor")
			os.Exit(1)
		}
		json.Unmarshal(parsedMonitor, &monitor)
		monitor.id = hit.(map[string]interface{})["_id"].(string)
		flippedDestinations := reverseMap(globalConfig.Destinations)

		for index := range monitor.Triggers {
			// Update destinationId and actioinId
			for k := range monitor.Triggers[index].Actions {
				destintionName := flippedDestinations[monitor.Triggers[index].Actions[k].DestinationID]
				if destintionName == "" {
					fmt.Println("Looks like remote monitor selected destination doesn't exists here, please update config")
					os.Exit(1)
				}
				monitor.Triggers[index].Actions[k].DestinationID = destintionName
			}
		}
		allMonitors = append(allMonitors, monitor)
	}
	allRemoteMonitorsMap = make(map[string]Monitor)
	remoteMonitorsSet := mapset.NewSet()
	for _, remoteMonitor := range allMonitors {
		remoteMonitorsSet.Add(remoteMonitor.Name)
		allRemoteMonitorsMap[remoteMonitor.Name] = remoteMonitor
	}
	return allRemoteMonitorsMap, remoteMonitorsSet
}

func prepareMonitor(localMonitor Monitor, remoteMonitor Monitor) Monitor {
	monitorToUpdate := localMonitor
	//Inject triggerIds in case updating existing triggers
	// Convert triggers to map
	remoteTriggers := make(map[string]Trigger)
	for _, remoteTrigger := range remoteMonitor.Triggers {
		remoteTriggers[remoteTrigger.Name] = remoteTrigger
	}
	//Update trigger if already existed
	// TODO::Same with Actions once released
	for index := range monitorToUpdate.Triggers {
		//Update trigger Id
		if remoteTrigger, ok := remoteTriggers[monitorToUpdate.Triggers[index].Name]; ok {
			monitorToUpdate.Triggers[index].ID = remoteTrigger.ID
		}
		// Update destinationId and actioinId
		for k := range monitorToUpdate.Triggers[index].Actions {
			destinationID := globalConfig.Destinations[monitorToUpdate.Triggers[index].Actions[k].DestinationID]
			if destinationID == "" {
				fmt.Println("destination specified doesn't exist in config file, verify it")
				os.Exit(1)
			}
			monitorToUpdate.Triggers[index].Actions[k].DestinationID = destinationID
		}
	}
	return monitorToUpdate
}

// TODO , check if the query is incorrect
func runMonitor(id string, monitor Monitor) bool {
	var r map[string]interface{}
	requestBody, err := json.Marshal(monitor)
	fmt.Println("requestBody", string(requestBody))
	resp, err := http.Post("http://localhost:9200/_opendistro/_alerting/monitors/_execute?dryrun=true", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		fmt.Println("Error retriving all the monitors", err)
		os.Exit(1)
	}
	defer resp.Body.Close()
	json.NewDecoder(resp.Body).Decode(&r)
	fmt.Println("r", r)
	res := r["trigger_results"].(map[string]interface{})
	executionResult, err := json.Marshal(res)
	var t interface{}
	err = json.Unmarshal(executionResult, &t)
	itemsMap := t.(map[string]interface{})
	for _, v := range itemsMap {
		var val map[string]interface{}
		asd, err := json.Marshal(v)
		if err != nil {
			fmt.Println("unable to find the proper response ")
			os.Exit(1)
		}
		json.Unmarshal(asd, &val)
		if val["error"] != nil {
			fmt.Println("Unable to run the monitor", val["error"])
			os.Exit(1)
		}
	}
	return true
}

func updateMonitor(remoteMonitor Monitor, monitor Monitor) {
	id := remoteMonitor.id
	var r map[string]interface{}
	client := http.Client{}
	a, err := json.Marshal(monitor)
	if err != nil {
		fmt.Println("Unable to parse monitor Object")
		os.Exit(1)
	}
	fmt.Println("Updating existing monitor", string(a))
	req, err := http.NewRequest(http.MethodPut, "http://localhost:9200/_opendistro/_alerting/monitors/"+id, bytes.NewBuffer(a))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error retriving all the monitors", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	json.NewDecoder(resp.Body).Decode(&r)
	fmt.Println(r)
}
