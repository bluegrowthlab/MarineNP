/*
 * MarineNP Data Models
 * Purpose: Define core data structures for the MarineNP database
 * Author: MarineNP Team
 * Date: 2025-06-10
 *
 * This file contains the core data models used throughout the MarineNP application,
 * defining the structure of molecules, organisms, properties, and related entities.
 */

package models

// Molecule represents a natural product compound with its structural and metadata information
type Molecule struct {
	ID                  int64     `json:"id" gorm:"primaryKey"`
	StandardInchi       string    `json:"standard_inchi"`
	StandardInchiKey    string    `json:"standard_inchi_key"`
	CanonicalSmiles     string    `json:"canonical_smiles"`
	SugarFreeSmiles     string    `json:"sugar_free_smiles"`
	Identifier          string    `json:"identifier"`
	Name                string    `json:"name"`
	Cas                 string    `json:"cas"`
	Synonyms            string    `json:"synonyms"`
	IupacName           string    `json:"iupac_name"`
	MurkoFramework      string    `json:"murko_framework"`
	StructuralComments  string    `json:"structural_comments"`
	NameTrustLevel      int       `json:"name_trust_level"`
	AnnotationLevel     int       `json:"annotation_level"`
	ParentID            int       `json:"parent_id"`
	VariantsCount       int       `json:"variants_count"`
	Ticker              int       `json:"ticker"`
	Status              string    `json:"status"`
	Active              bool      `json:"active"`
	HasVariants         bool      `json:"has_variants"`
	HasStereo           bool      `json:"has_stereo"`
	IsTautomer          bool      `json:"is_tautomer"`
	IsParent            bool      `json:"is_parent"`
	IsPlaceholder       bool      `json:"is_placeholder"`
	Comment             string    `json:"comment"`
	CreatedAt           SQLiteTime `json:"created_at"`
	UpdatedAt           SQLiteTime `json:"updated_at"`
	OrganismCount       int       `json:"organism_count"`
	GeoCount            int       `json:"geo_count"`
	CitationCount       int       `json:"citation_count"`
	CollectionCount     int       `json:"collection_count"`
	SynonymCount        int       `json:"synonym_count"`
	IsDuplicate         bool      `json:"is_duplicate"`
	Properties          Properties `json:"properties" gorm:"foreignKey:MoleculeID"`
	Organisms           []Organism `json:"organisms" gorm:"many2many:molecule_organism;"`
	GeoLocations        []GeoLocation `json:"geo_locations" gorm:"many2many:geo_location_molecule;"`
}

// Organism represents a biological organism with taxonomic and metadata information
type Organism struct {
	ID                  int64     `json:"id" gorm:"primaryKey"`
	Name                string    `json:"name"`
	IRI                 string    `json:"iri"`
	CreatedAt           SQLiteTime `json:"created_at"`
	UpdatedAt           SQLiteTime `json:"updated_at"`
	Rank                string    `json:"rank"`
	MoleculeCount       int       `json:"molecule_count" gorm:"default:0"`
	Slug                string    `json:"slug"`
	AphiaIDWorms        *int      `json:"aphiaid_worms" gorm:"column:aphiaid_worms"`
	NameAphiaWorms      string    `json:"name_aphia_worms"`
	EnvironmentAphiaWorms string  `json:"environment_aphia_worms"`
	IsMarine            *bool     `json:"is_marine" gorm:"default:false"`
	Molecules           []Molecule `json:"molecules" gorm:"many2many:molecule_organism;"`
}

// Properties represents chemical properties and descriptors of a molecule
type Properties struct {
	ID                              int64     `json:"id" gorm:"primaryKey"`
	MoleculeID                      int64     `json:"molecule_id"`
	TotalAtomCount                  int       `json:"total_atom_count" gorm:"default:0"`
	HeavyAtomCount                  int       `json:"heavy_atom_count" gorm:"default:0"`
	MolecularWeight                 float64   `json:"molecular_weight" gorm:"default:0"`
	ExactMolecularWeight            float64   `json:"exact_molecular_weight" gorm:"default:0"`
	MolecularFormula                string    `json:"molecular_formula"`
	Alogp                           float64   `json:"alogp" gorm:"type:numeric(8,2);default:0"`
	TopologicalPolarSurfaceArea     float64   `json:"topological_polar_surface_area" gorm:"type:numeric(8,2);default:0"`
	RotatableBondCount              int       `json:"rotatable_bond_count" gorm:"default:0"`
	HydrogenBondAcceptors           int       `json:"hydrogen_bond_acceptors" gorm:"default:0"`
	HydrogenBondDonors              int       `json:"hydrogen_bond_donors" gorm:"default:0"`
	HydrogenBondAcceptorsLipinski   int       `json:"hydrogen_bond_acceptors_lipinski" gorm:"default:0"`
	HydrogenBondDonorsLipinski      int       `json:"hydrogen_bond_donors_lipinski" gorm:"default:0"`
	LipinskiRuleOfFiveViolations    int       `json:"lipinski_rule_of_five_violations"`
	AromaticRingsCount              int       `json:"aromatic_rings_count" gorm:"default:0"`
	QEDDrugLikeliness              float64   `json:"qed_drug_likeliness" gorm:"type:numeric(8,2);default:0"`
	FormalCharge                    int       `json:"formal_charge" gorm:"default:0"`
	FractionCSP3                    float64   `json:"fractioncsp3" gorm:"type:numeric(8,2);default:0"`
	NumberOfMinimalRings            int       `json:"number_of_minimal_rings"`
	VanDerWallsVolume              float64   `json:"van_der_walls_volume" gorm:"type:numeric(8,2)"`
	ContainsSugar                   bool      `json:"contains_sugar"`
	ContainsRingSugars              bool      `json:"contains_ring_sugars"`
	ContainsLinearSugars            bool      `json:"contains_linear_sugars"`
	Fragments                       string    `json:"fragments" gorm:"type:jsonb"`
	FragmentsWithSugar              string    `json:"fragments_with_sugar" gorm:"type:jsonb"`
	MurckoFramework                 string    `json:"murcko_framework"`
	NPLikeness                      float64   `json:"np_likeness" gorm:"type:numeric(8,2)"`
	ChemicalClass                   string    `json:"chemical_class"`
	ChemicalSubClass                string    `json:"chemical_sub_class"`
	ChemicalSuperClass              string    `json:"chemical_super_class"`
	DirectParentClassification      string    `json:"direct_parent_classification"`
	CreatedAt                       SQLiteTime `json:"created_at"`
	UpdatedAt                       SQLiteTime `json:"updated_at"`
	NPClassifierPathway             string    `json:"np_classifier_pathway"`
	NPClassifierSuperclass          string    `json:"np_classifier_superclass"`
	NPClassifierClass               string    `json:"np_classifier_class"`
	NPClassifierIsGlycoside         bool      `json:"np_classifier_is_glycoside"`
}

// Collection represents a curated collection of molecules with metadata
type Collection struct {
	ID               int64     `json:"id" gorm:"primaryKey"`
	Title            string    `json:"title"`
	Slug             string    `json:"slug"`
	Description      string    `json:"description"`
	Comments         string    `json:"comments"`
	Identifier       string    `json:"identifier"`
	URL              string    `json:"url"`
	Photo            string    `json:"photo"`
	IsPublic         bool      `json:"is_public"`
	UUID             string    `json:"uuid"`
	Status           string    `json:"status"`
	JobsStatus       string    `json:"jobs_status"`
	JobInfo          string    `json:"job_info"`
	DOI              string    `json:"doi"`
	OwnerID          int64     `json:"owner_id"`
	LicenseID        int64     `json:"license_id"`
	ReleaseDate      SQLiteTime `json:"release_date"`
	CreatedAt        SQLiteTime `json:"created_at"`
	UpdatedAt        SQLiteTime `json:"updated_at"`
	Image            string    `json:"image"`
	Promote          bool      `json:"promote"`
	SortOrder        int64     `json:"sort_order"`
	SuccessfulEntries int64    `json:"successful_entries"`
	FailedEntries    int64     `json:"failed_entries"`
	MoleculesCount   int64     `json:"molecules_count"`
	CitationsCount   int64     `json:"citations_count"`
	OrganismsCount   int64     `json:"organisms_count"`
	GeoCount         int64     `json:"geo_count"`
	TotalEntries     int64     `json:"total_entries"`
	Molecules        []Molecule `json:"molecules" gorm:"many2many:collection_molecule;"`
}

// Citation represents a literature reference for a molecule
type Citation struct {
	ID           int64     `json:"id" gorm:"primaryKey"`
	DOI          string    `json:"doi"`
	Title        string    `json:"title"`
	Authors      string    `json:"authors"`
	CitationText string    `json:"citation_text"`
	Active       bool      `json:"active"`
	CreatedAt    SQLiteTime `json:"created_at"`
	UpdatedAt    SQLiteTime `json:"updated_at"`
}

// GeoLocation represents a geographic location where molecules were found
type GeoLocation struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	CreatedAt SQLiteTime `json:"created_at"`
	UpdatedAt SQLiteTime `json:"updated_at"`
	Molecules []Molecule `json:"molecules" gorm:"many2many:geo_location_molecule;"`
}

// OBISCache represents cached data from the Ocean Biogeographic Information System
type OBISCache struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	AphiaIDWorms int      `json:"aphiaid_worms" gorm:"column:aphiaid_worms;uniqueIndex"`
	OBISData    string    `json:"obis_data" gorm:"type:jsonb;column:obis_data"`
	CreatedAt   SQLiteTime `json:"created_at"`
	UpdatedAt   SQLiteTime `json:"updated_at"`
}

// TableName specifies the table name for OBISCache
func (OBISCache) TableName() string {
	return "obis_cache"
} 