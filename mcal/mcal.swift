import Foundation
import EventKit

setbuf(stdout, nil)

var eventDuration = 60.0 // Default event duration in minutes
let eventStore = EKEventStore()

// Entry point
func main() {
    handleCommandLineArguments()
    checkCalendarAuthorizationStatus()
    
    let currentDate = Date()
    let startOfDay = Calendar.current.startOfDay(for: currentDate)
    let tomorrow = Calendar.current.date(byAdding: .day, value: 1, to: startOfDay)!
    
    let calendars = eventStore.calendars(for: .event)
    
    let allEvents = fetchEvents(start: startOfDay, end: tomorrow, calendars: calendars)
    
    displayEvents(events: allEvents)
}

// Handle command-line arguments
func handleCommandLineArguments() {
    let arguments = CommandLine.arguments
    
    guard arguments.count > 1 else {
        print("Invalid input. Please provide a valid command.")
        exit(1)
    }
    
    switch arguments[1] {
    case "l", "list":
        break
    default:
        if arguments.count > 2 {
            eventDuration = Double(arguments[2]) ?? 60.0
        } else {
            print("Invalid input. Please provide a valid command.")
            exit(1)
        }
    }
}

// Check calendar authorization status
func checkCalendarAuthorizationStatus() {
    switch EKEventStore.authorizationStatus(for: .event) {
    case .authorized:
        break
    case .denied:
        print("Access denied. Please go to Settings > Privacy & Security > Calendars > Terminal.")
        exit(1)
    case .notDetermined:
        eventStore.requestFullAccessToEvents { granted, error in
            if granted {
                print("Access granted.")
            } else {
                print("Access denied.")
            }
        }
    default:
        fputs("Unknown authorization status", stderr)
    }
}

// Fetch events from the calendar
func fetchEvents(start: Date, end: Date, calendars: [EKCalendar]) -> [EKEvent] {
    let events = eventStore.events(matching: eventStore.predicateForEvents(withStart: start, end: end, calendars: calendars))
    return events.filter { !$0.isAllDay }
}

// Display events with simplified output
func displayEvents(events: [EKEvent]) {
    for event in events {
        // Clean up the title by removing email addresses and the word "Work"
        let cleanedTitle = event.title!.replacingOccurrences(of: "Work", with: "")
                                .replacingOccurrences(of: "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}", with: "", options: .regularExpression)

        // Location and separator
        let locationInfo = event.location != nil ? " | \(event.location!)" : ""
        let eventTime = event.startDate.format(f: "HH:mm")
        
        print("\(eventTime) \(cleanedTitle)\(locationInfo)")
    }
}

// Date extension for formatting
extension Date {
    func format(f: String) -> String {
        let dateFormatter = DateFormatter()
        dateFormatter.dateFormat = f
        return dateFormatter.string(from: self)
    }
}

// Execute the main function
main()
