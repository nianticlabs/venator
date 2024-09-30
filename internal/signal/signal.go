package signal

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/nianticlabs/venator/internal/config"
)

// Struct for the output Signal
type Signal struct {
	Timestamp        time.Time           `json:"timestamp"`
	Rule_ID          string              `json:"rule_id"`
	Rule_Name        string              `json:"rule_name"`
	ConfidenceID     int                 `json:"confidenceid"`
	Confidence       string              `json:"confidence"`
	TTPs             []map[string]string `json:"ttps"`
	Actor            Actor               `json:"actor"`
	Resource         Resource            `json:"resource"`
	SrcEndpoint      Endpoint            `json:"src_endpoint"`
	DstEndpoint      Endpoint            `json:"dst_endpoint"`
	Message          string              `json:"message"`
	Metadata         Metadata            `json:"metadata"`
	RuleSpecificData map[string]string   `json:"rule_specific_data"`
}

type Actor struct {
	User User `json:"user"`
}

type User struct {
	Name string `json:"name"`
	UID  string `json:"uid"`
}

// Resource that was affected by the activity/event (i.e. target of the activity)
type Resource struct {
	Name string `json:"name"`
	Type string `json:"type"`
	UID  string `json:"uid"`
}

// Endpoint represents a network endpoint
type Endpoint struct {
	Hostname string `json:"hostname"`
	IP       string `json:"ip"`
}

// Metadata contains additional information about the event
type Metadata struct {
	EventID    string `json:"event_id"`
	EventIndex string `json:"event_index"`
}

const (
	ConfidenceUnknown int = 0
	ConfidenceLow     int = 1
	ConfidenceMedium  int = 2
	ConfidenceHigh    int = 3
)

func BuildSignal(result map[string]string, cfg *config.RuleConfig) (*Signal, error) {
	if len(result) < len(cfg.Output.Fields) {
		return nil, fmt.Errorf("number of query result fields mismatches expected count")
	}
	signal := Signal{
		Rule_ID:      cfg.UID,
		Rule_Name:    cfg.Name,
		ConfidenceID: getConfidenceID(cfg.Confidence),
		Confidence:   string(cfg.Confidence),
		TTPs:         []map[string]string{},
	}
	for _, ttp := range cfg.TTPs {
		signal.TTPs = append(signal.TTPs, map[string]string{
			"framework": ttp.Framework,
			"tactic":    ttp.Tactic,
			"name":      ttp.Name,
			"id":        ttp.ID,
		})
	}

	for _, outputField := range cfg.Output.Fields {
		value, exists := result[outputField.Source]
		if !exists {
			return nil, fmt.Errorf("source field %s not found in query results", outputField.Source)
		}
		switch outputField.Field {
		case "Timestamp":
			parsedTime, err := time.Parse(time.RFC3339, value)
			if err != nil {
				return nil, fmt.Errorf("failed to parse timestamp: %w", err)
			}
			signal.Timestamp = parsedTime
		case "ActorUserName":
			signal.Actor.User.Name = value
		case "ActorUserUID":
			signal.Actor.User.UID = value
		case "ResourceName":
			signal.Resource.Name = value
		case "ResourceType":
			signal.Resource.Type = value
		case "ResourceUID":
			signal.Resource.UID = value
		case "SrcHostname":
			signal.SrcEndpoint.Hostname = value
		case "SrcIP":
			signal.SrcEndpoint.IP = value
		case "DstHostname":
			signal.DstEndpoint.Hostname = value
		case "DstIP":
			signal.DstEndpoint.IP = value
		case "Message":
			signal.Message = value
		case "EventID":
			signal.Metadata.EventID = value
		case "EventIndex":
			signal.Metadata.EventIndex = value
		case "RuleSpecificData":
			var rsd map[string]interface{}
			if err := json.Unmarshal([]byte(value), &rsd); err == nil {
				signal.RuleSpecificData = make(map[string]string)
				for k, v := range rsd {
					signal.RuleSpecificData[k] = fmt.Sprintf("%v", v)
				}
			} else {
				signal.RuleSpecificData = map[string]string{"raw": value}
			}

		default:
			return nil, fmt.Errorf("unsupported output field %s", outputField.Field)
		}
	}
	return &signal, nil
}

func BuildOutput(result map[string]string, cfg *config.RuleConfig) (any, error) {
	var output any
	switch cfg.Output.Format {
	case config.OutputFormatSignal:
		sig, err := BuildSignal(result, cfg)
		if err != nil {
			return nil, err
		}
		output = sig
	case config.OutputFormatRaw:
		output = result
	default:
		return nil, fmt.Errorf("unsupported output format: %s", cfg.Output.Format)
	}
	return output, nil
}

func getConfidenceID(confidence config.ConfidenceLevel) int {
	switch confidence {
	case config.ConfidenceLow:
		return ConfidenceLow
	case config.ConfidenceMedium:
		return ConfidenceMedium
	case config.ConfidenceHigh:
		return ConfidenceHigh
	default:
		return ConfidenceUnknown
	}
}
