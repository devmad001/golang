package main

import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"
    "time"

    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

// Models
type Patient struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string            `json:"name" bson:"name"`
    Email       string            `json:"email" bson:"email"`
    Age         int               `json:"age" bson:"age"`
    Gender      string            `json:"gender" bson:"gender"`
    BloodGroup  string            `json:"bloodGroup" bson:"bloodGroup"`
    ContactNo   string            `json:"contactNo" bson:"contactNo"`
    CreatedAt   time.Time         `json:"createdAt" bson:"createdAt"`
}

type Doctor struct {
    ID           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name         string            `json:"name" bson:"name"`
    Email        string            `json:"email" bson:"email"`
    Specialization string          `json:"specialization" bson:"specialization"`
    Department    string           `json:"department" bson:"department"`
    ContactNo    string            `json:"contactNo" bson:"contactNo"`
    CreatedAt    time.Time         `json:"createdAt" bson:"createdAt"`
}

type Appointment struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    PatientID   primitive.ObjectID `json:"patientId" bson:"patientId"`
    DoctorID    primitive.ObjectID `json:"doctorId" bson:"doctorId"`
    DateTime    time.Time          `json:"dateTime" bson:"dateTime"`
    Status      string            `json:"status" bson:"status"` // Scheduled, Completed, Cancelled
    Description string            `json:"description" bson:"description"`
    CreatedAt   time.Time         `json:"createdAt" bson:"createdAt"`
}

type Department struct {
    ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
    Name        string            `json:"name" bson:"name"`
    Description string            `json:"description" bson:"description"`
    CreatedAt   time.Time         `json:"createdAt" bson:"createdAt"`
}

// Database collections
var (
    client *mongo.Client
    patientCollection *mongo.Collection
    doctorCollection *mongo.Collection
    appointmentCollection *mongo.Collection
    departmentCollection *mongo.Collection
)

func init() {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
    var err error
    
    client, err = mongo.Connect(ctx, clientOptions)
    if err != nil {
        log.Fatal(err)
    }

    err = client.Ping(ctx, nil)
    if err != nil {
        log.Fatal(err)
    }

    fmt.Println("Connected to MongoDB!")

    // Initialize collections
    db := client.Database("hospitaldb")
    patientCollection = db.Collection("patients")
    doctorCollection = db.Collection("doctors")
    appointmentCollection = db.Collection("appointments")
    departmentCollection = db.Collection("departments")

    // Create indexes
    createIndexes(ctx)
}

func createIndexes(ctx context.Context) {
    // Patient email index
    patientIndex := mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    _, err := patientCollection.Indexes().CreateOne(ctx, patientIndex)
    if err != nil {
        log.Printf("Error creating patient index: %v\n", err)
    }

    // Doctor email index
    doctorIndex := mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    _, err = doctorCollection.Indexes().CreateOne(ctx, doctorIndex)
    if err != nil {
        log.Printf("Error creating doctor index: %v\n", err)
    }
}

// Patient handlers
func createPatient(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var patient Patient
    if err := json.NewDecoder(r.Body).Decode(&patient); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    patient.CreatedAt = time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := patientCollection.InsertOne(ctx, patient)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    patient.ID = result.InsertedID.(primitive.ObjectID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(patient)
}

func getPatients(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    cursor, err := patientCollection.Find(ctx, bson.M{})
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer cursor.Close(ctx)

    var patients []Patient
    if err = cursor.All(ctx, &patients); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(patients)
}

// Doctor handlers
func createDoctor(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var doctor Doctor
    if err := json.NewDecoder(r.Body).Decode(&doctor); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    doctor.CreatedAt = time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := doctorCollection.InsertOne(ctx, doctor)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    doctor.ID = result.InsertedID.(primitive.ObjectID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(doctor)
}

// Appointment handlers
func createAppointment(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var appointment Appointment
    if err := json.NewDecoder(r.Body).Decode(&appointment); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    appointment.CreatedAt = time.Now()
    appointment.Status = "Scheduled"
    
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // Validate patient and doctor existence
    if err := validateAppointment(ctx, &appointment); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    result, err := appointmentCollection.InsertOne(ctx, appointment)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    appointment.ID = result.InsertedID.(primitive.ObjectID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(appointment)
}

func validateAppointment(ctx context.Context, appointment *Appointment) error {
    // Check if patient exists
    var patient Patient
    err := patientCollection.FindOne(ctx, bson.M{"_id": appointment.PatientID}).Decode(&patient)
    if err != nil {
        return fmt.Errorf("patient not found")
    }

    // Check if doctor exists
    var doctor Doctor
    err = doctorCollection.FindOne(ctx, bson.M{"_id": appointment.DoctorID}).Decode(&doctor)
    if err != nil {
        return fmt.Errorf("doctor not found")
    }

    return nil
}

// Department handlers
func createDepartment(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var department Department
    if err := json.NewDecoder(r.Body).Decode(&department); err != nil {
        http.Error(w, err.Error(), http.StatusBadRequest)
        return
    }

    department.CreatedAt = time.Now()
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    result, err := departmentCollection.InsertOne(ctx, department)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    department.ID = result.InsertedID.(primitive.ObjectID)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(department)
}

func main() {
    defer func() {
        if client != nil {
            if err := client.Disconnect(context.Background()); err != nil {
                log.Printf("Error disconnecting from MongoDB: %v\n", err)
            }
        }
    }()

    // Patient routes
    http.HandleFunc("/patients", createPatient)
    http.HandleFunc("/patients/list", getPatients)

    // Doctor routes
    http.HandleFunc("/doctors", createDoctor)

    // Appointment routes
    http.HandleFunc("/appointments", createAppointment)

    // Department routes
    http.HandleFunc("/departments", createDepartment)

    fmt.Println("Starting hospital management service on http://localhost:8080")
    if err := http.ListenAndServe(":8080", nil); err != nil {
        fmt.Printf("Error starting server: %v\n", err)
    }
} 